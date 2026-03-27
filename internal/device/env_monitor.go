package device

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/go-routeros/routeros"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type EnvCollector struct {
	db *gorm.DB
}

func NewEnvCollector(db *gorm.DB) *EnvCollector {
	return &EnvCollector{db: db}
}

func (c *EnvCollector) CollectAllDevices(ctx context.Context) error {
	var nasList []domain.NetNas
	if err := c.db.Where("status = ?", "enabled").Find(&nasList).Error; err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, nas := range nasList {
		wg.Add(1)
		go func(nas domain.NetNas) {
			defer wg.Done()
			if err := c.CollectDevice(ctx, &nas); err != nil {
				zap.S().Errorw("Failed to collect env metrics", "nas_id", nas.ID, "error", err)
			}
		}(nas)
	}
	wg.Wait()
	return nil
}

func (c *EnvCollector) CollectDevice(ctx context.Context, nas *domain.NetNas) error {
	apiUser := nas.ApiUser
	apiPass := nas.ApiPass
	if apiUser == "" {
		apiUser = "admin"
	}
	if apiPass == "" {
		apiPass = nas.Secret
	}

	zap.S().Debugw("Connecting to MikroTik for env metrics", "nas_id", nas.ID, "ip", nas.Ipaddr, "user", apiUser)
	
	address := nas.Ipaddr + ":8728"
	client, err := routeros.Dial(address, apiUser, apiPass)
	if err != nil {
		zap.S().Errorw("Failed to connect to MikroTik", "nas_id", nas.ID, "ip", nas.Ipaddr, "error", err)
		return fmt.Errorf("failed to connect to %s: %w", nas.Ipaddr, err)
	}
	defer client.Close()

	re, err := client.Run("/system/health/print")
	if err != nil {
		zap.S().Debugw("No /system/health available", "nas_id", nas.ID, "error", err)
	} else {
		zap.S().Debugw("Got system health", "nas_id", nas.ID, "response", re.Re)
	}

	now := time.Now()
	var metrics []domain.EnvironmentMetric

	if temp, err := c.getTemperature(client); err == nil && temp > 0 {
		metrics = append(metrics, domain.EnvironmentMetric{
			TenantID:    fmt.Sprintf("%d", nas.TenantID),
			NasID:       uint(nas.ID),
			NasName:     nas.Name,
			MetricType:  domain.MetricTypeTemperature,
			Value:       temp,
			Unit:        "C",
			Severity:    c.calculateSeverity(domain.MetricTypeTemperature, temp),
			CollectedAt: now,
			CreatedAt:   now,
		})
	}

	if power, err := c.getPower(client); err == nil && power > 0 {
		metrics = append(metrics, domain.EnvironmentMetric{
			TenantID:    fmt.Sprintf("%d", nas.TenantID),
			NasID:       uint(nas.ID),
			NasName:     nas.Name,
			MetricType:  domain.MetricTypePower,
			Value:       power,
			Unit:        "W",
			Severity:    c.calculateSeverity(domain.MetricTypePower, power),
			CollectedAt: now,
			CreatedAt:   now,
		})
	}

	if voltage, err := c.getVoltage(client); err == nil && voltage > 0 {
		metrics = append(metrics, domain.EnvironmentMetric{
			TenantID:    fmt.Sprintf("%d", nas.TenantID),
			NasID:       uint(nas.ID),
			NasName:     nas.Name,
			MetricType:  domain.MetricTypeVoltage,
			Value:       voltage,
			Unit:        "V",
			Severity:    c.calculateSeverity(domain.MetricTypeVoltage, voltage),
			CollectedAt: now,
			CreatedAt:   now,
		})
	}

	if fan, err := c.getFanSpeed(client); err == nil && fan > 0 {
		metrics = append(metrics, domain.EnvironmentMetric{
			TenantID:    fmt.Sprintf("%d", nas.TenantID),
			NasID:       uint(nas.ID),
			NasName:     nas.Name,
			MetricType:  domain.MetricTypeFanSpeed,
			Value:       fan,
			Unit:        "RPM",
			Severity:    c.calculateSeverity(domain.MetricTypeFanSpeed, fan),
			CollectedAt: now,
			CreatedAt:   now,
		})
	}

	if len(metrics) > 0 {
		zap.S().Infow("Saving environment metrics", "nas_id", nas.ID, "count", len(metrics), "metrics", metrics)
		if err := c.db.Create(&metrics).Error; err != nil {
			zap.S().Errorw("Failed to save metrics", "nas_id", nas.ID, "error", err)
			return err
		}
		zap.S().Infow("Metrics saved successfully", "nas_id", nas.ID)
	}
	return nil
}

func (c *EnvCollector) getTemperature(client *routeros.Client) (float64, error) {
	re, err := client.Run("/system/health/print")
	if err == nil && len(re.Re) > 0 {
		temp := getFloatFromSentence(re.Re[0], "cpu-temperature")
		zap.S().Debugw("Temperature check", "cpu-temperature", temp)
		if temp > 0 {
			return temp, nil
		}
		temp = getFloatFromSentence(re.Re[0], "temperature")
		zap.S().Debugw("Temperature check", "temperature", temp)
		if temp > 0 {
			return temp, nil
		}
	}
	re, err = client.Run("/system/resource/print")
	if err == nil && len(re.Re) > 0 {
		temp := getFloatFromSentence(re.Re[0], "cpu-temperature")
		zap.S().Debugw("Temperature check from resource", "cpu-temperature", temp)
		if temp > 0 {
			return temp, nil
		}
	}
	return 0, fmt.Errorf("temperature not available")
}

func (c *EnvCollector) getPower(client *routeros.Client) (float64, error) {
	re, err := client.Run("/system/health/print")
	if err != nil {
		return 0, err
	}
	if len(re.Re) > 0 {
		power := getFloatFromSentence(re.Re[0], "power-consumption")
		zap.S().Debugw("Power check", "power", power)
		return power, nil
	}
	return 0, fmt.Errorf("power not available")
}

func (c *EnvCollector) getVoltage(client *routeros.Client) (float64, error) {
	re, err := client.Run("/system/health/print")
	if err != nil {
		return 0, err
	}
	if len(re.Re) > 0 {
		volt := getFloatFromSentence(re.Re[0], "voltage")
		zap.S().Debugw("Voltage check", "voltage", volt)
		if volt > 0 {
			return volt, nil
		}
	}
	return 0, fmt.Errorf("voltage not available")
}

func (c *EnvCollector) getFanSpeed(client *routeros.Client) (float64, error) {
	re, err := client.Run("/system/health/print")
	if err != nil {
		return 0, err
	}
	if len(re.Re) > 0 {
		return getFloatFromSentence(re.Re[0], "fan1-speed"), nil
	}
	return 0, fmt.Errorf("fan speed not available")
}

func getFloatFromSentence(sentence interface{}, key string) float64 {
	type sentenceGetter interface {
		GetList() []interface{}
	}
	
	if sg, ok := sentence.(sentenceGetter); ok {
		list := sg.GetList()
		for _, item := range list {
			if attr, ok := item.(interface{ GetKey() string; GetValue() string }); ok {
				if attr.GetKey() == key {
					var f float64
					val := attr.GetValue()
					fmt.Sscanf(val, "%f", &f)
					return f
				}
			}
		}
	}
	
	v := reflect.ValueOf(sentence)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		mapField := v.FieldByName("Map")
		if mapField.IsValid() && !mapField.IsZero() {
			zap.S().Debugw("Found Map field", "type", mapField.Type().String())
			if mapStr, ok := mapField.Interface().(map[string]string); ok {
				zap.S().Debugw("Map is map[string]string", "map", mapStr)
				if val, ok := mapStr[key]; ok {
					var f float64
					fmt.Sscanf(val, "%f", &f)
					return f
				}
			}
		}
	}
	return 0
}

func (c *EnvCollector) calculateSeverity(metricType string, value float64) string {
	switch metricType {
	case domain.MetricTypeTemperature:
		if value >= 80 {
			return domain.SeverityCritical
		}
		if value >= 70 {
			return domain.SeverityWarning
		}
		return domain.SeverityNormal

	case domain.MetricTypePower:
		if value >= 120 {
			return domain.SeverityCritical
		}
		if value >= 100 {
			return domain.SeverityWarning
		}
		return domain.SeverityNormal

	case domain.MetricTypeVoltage:
		if value >= 260 || value < 180 {
			return domain.SeverityCritical
		}
		if value >= 250 || value < 200 {
			return domain.SeverityWarning
		}
		return domain.SeverityNormal

	case domain.MetricTypeFanSpeed:
		if value < 1000 {
			return domain.SeverityWarning
		}
		return domain.SeverityNormal

	default:
		return domain.SeverityNormal
	}
}
