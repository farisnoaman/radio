package maintenance

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaintenanceManager_Toggle(t *testing.T) {
	// Since we haven't mocked GORM yet and NewMaintenanceManager requires it,
	// checking if we can test just the logic or if we need to mock DB.
	// The Enable/Disable/IsActive methods don't use DB.
	
	m := &MaintenanceManager{}
	
	assert.False(t, m.IsActive())

	err := m.Enable()
	assert.NoError(t, err)
	assert.True(t, m.IsActive())

	err = m.Disable()
	assert.NoError(t, err)
	assert.False(t, m.IsActive())
}

func TestMaintenanceManager_Concurrency(t *testing.T) {
	m := &MaintenanceManager{}
	var wg sync.WaitGroup
	
	// Concurrent reads (IsActive)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.IsActive()
		}()
	}

	// Concurrent writes (Enable/Disable)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Enable()
			m.Disable()
		}()
	}
	
	wg.Wait()
}
