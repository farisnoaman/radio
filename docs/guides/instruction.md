faris@faris:~/Downloads/toughradius/toughradius$ 
faris@faris:~/Downloads/toughradius/toughradius$ ^C
faris@faris:~/Downloads/toughradius/toughradius$ go version && node -v && npm -v
Command 'go' not found, but can be installed with:
faris@faris:~/Downloads/toughradius/toughradius$ ^C
faris@faris:~/Downloads/toughradius/toughradius$ node -v && npm -v
v22.12.0
10.9.0
faris@faris:~/Downloads/toughradius/toughradius$ u
name -m
x86_64
faris@faris:~/Downloads/toughradius/toughradius$ uname -m
x86_64
faris@faris:~/Downloads/toughradius/toughradius$ 
       curl -L https://go.dev/dl/go1.24.0.linux-amfaris@faris:~/Downloads/toughradius/toughradius$         mkdir -p $HOME/go;
d64.tar.gz -o go1.24.0.linux-amd64.tar.gz;
        tar -xxvf go1.24.0.linux-amd64.tar.gz;
        exfaris@faris:~/Downloads/toughradius/toughradius$         cd $HOME/go;
port PATH=$HOME/go/go/bin:$PATH;
        go version
    
faris@faris:~/go$         curl -L https://go.dev/dl/go1.24.0.linux-amd64.tar.gz -o go1.24.0.linux-amd64.tar.gz;
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--  0     0    0     0    0     0      0      0 --:--100    75  100    75    0     0     66      0  0:00:01  0:00:01 --:--:--    66
  0     0    0     0    0     0      0      0 --:  0 74.7M    0  338k    0     0   125k      0  0:  2 74.7M    2 1855k    0     0   490k      0  0:  4 74.7M    4 3295k    0     0   702k      0  0:  6 74.7M    6 4751k    0     0   834k      0  0:  7 74.7M    7 6111k    0     0   912k      0  0:01:23  0:00:06  0:01:17 1227k^C
faris@faris:~/go$ cd /home/faris/Downloads/toughradius/toughradius
faris@faris:~/Downloads/toughradius/toughradius$
64.tar.gz;
        tar -C $HOME/go -xzf go1.24.0.lfaris@fari        wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz;
inux-amd64.tar.gz;
        export PATH=$HOME/go/go/bin:$PATH;
        go version;
    
--2026-02-16 15:46:04--  https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
Resolving go.dev (go.dev)... 216.239.38.21, 216.239.34.21, 216.239.32.21, ...
Connecting to go.dev (go.dev)|216.239.38.21|:443... connected.
HTTP request sent, awaiting response... 302 Found
Location: https://dl.google.com/go/go1.24.0.linux-amd64.tar.gz [following]
--2026-02-16 15:46:04--  https://dl.google.com/go/go1.24.0.linux-amd64.tar.gz
Resolving dl.google.com (dl.google.com)... 192.178.54.46, 2a00:1450:401a:801::200e
Connecting to dl.google.com (dl.google.com)|192.178.54.46|:443... connected.
HTTP request sent, awaiting response... 200 OK
Length: 78382844 (75M) [application/x-gzip]
Saving to: ‘go1.24.0.linux-amd64.tar.gz’

      go1.24   0%       0  --.-KB/s                                                                                      go1.24.   0%  19.64K  84.6KB/s                                                                                     go1.24.0   0%  51.94K   114KB/s                                                                                    go1.24.0.   0%  95.00K   137KB/s                                                                                   go1.24.0.l   0% 159.60K   174KB/s                                                                                  go1.24.0.li   0% 234.96K   203KB/s                                                                                 go1.24.0.lin   0% 358.27K   256KB/s                                                                                 o1.24.0.linu   0% 503.61K   310KB/s                                                                                 1.24.0.linux   0% 738.08K   404KB/s                                                                                 .24.0.linux-   1% 944.29K   465KB/s                                                                                 24.0.linux-a   1%   1.33M   608KB/s                                                                                 4.0.linux-am   2%   1.62M   682KB/s                                                                                 .0.linux-amd   2%   1.64M   574KB/s                                                                                 0.linux-amd6   2%   1.70M   535KB/s    eta 2m 20s                                                                .linux-amd64   3%   2.69M   789KB/s    eta 2m 20s                                                            linux-amd64.   4%   3.20M   889KB/s    eta 2m 20s                                                        inux-amd64.t   4%   3.53M   927KB/s    eta 2m 20s                                                    nux-amd64.ta   5%   3.86M  1.02MB/s    eta 2m 20s                                                ux-amd64.tar   5%   4.19M  1.09MB/s    eta 73s                                               x-amd64.tar.   5%   4.47M  1.14MB/s    eta 73s                                           -amd64.tar.g   6%   4.58M  1.05MB/s    eta 73s                                       amd64.tar.gz   6%   4.86M  1.10MB/s    eta 77s                                   md64.tar.gz    7%   5.91M  1.30MB/s    eta 77s                               d64.tar.gz     8%   6.14M  1.32MB/s    eta 77s                           64.tar.gz      8%   6.16M  1.22MB/s    eta 77s                       4.tar.gz       8%   6.72M  1.30MB/s    eta 65s                   .tar.gz        9%   6.87M  1.28MB/s    eta 65s               tar.gz         9%   7.05M  1.21MB/s    eta 65s           ar.gz          9%   7.12M  1.15MB/s    eta 65s       r.gz           9%   7.34M  1.24MB/s    eta 69s   .gz           10%   7.78M  1.35MB/s    eta 69s                                                                      gz            10%   7.98M  1.18MB/s    eta 69s                                                                      z             10%   8.19M  1.12MB/s    eta 69s                                                                                    11%   8.47M  1.10MB/s    eta 69s                                                                                 g  11%   8.67M  1.08MB/s    eta 65s                                                                                go  11%   8.90M  1.06MB/s    eta 65s                                                                               go1  12%   9.12M  1.06MB/s    eta 65s                                                                              go1.  12%   9.43M  1.21MB/s    eta 65s                                                                             go1.2  12%   9.66M  1.20MB/s    eta 65s                                                                            go1.24  13%   9.95M  1.19MB/s    eta 62s                                                                           go1.24.  13%  10.25M  1.16MB/s    eta 62s                                                                          go1.24.0  14%  10.53M  1.08MB/s    eta 62s                                                                         go1.24.0.  14%  10.80M  1.11MB/s    eta 62s                                                                        go1.24.0.l  14%  11.09M  1.16MB/s    eta 62s                                                                    go1.24.0.li  15%  11.44M  1.30MB/s    eta 59s                                                               go1.24.0.lin  15%  11.66M  1.16MB/s    eta 59s                                                           o1.24.0.linu  16%  12.00M  1.20MB/s    eta 59s                                                       1.24.0.linux  16%  12.34M  1.24MB/s    eta 59s                                                   .24.0.linux-  16%  12.59M  1.25MB/s    eta 58s                                               24.0.linux-a  17%  12.76M  1.17MB/s    eta 58s                                           4.0.linux-am  17%  13.11M  1.19MB/s    eta 58s                                       .0.linux-amd  17%  13.37M  1.20MB/s    eta 58s                                   0.linux-amd6  18%  13.66M  1.19MB/s    eta 57s                               .linux-amd64  18%  13.80M  1.17MB/s    eta 57s                           linux-amd64.  18%  14.17M  1.18MB/s    eta 57s                       inux-amd64.t  19%  14.45M  1.18MB/s    eta 57s                   nux-amd64.ta  19%  14.70M  1.16MB/s    eta 57s               ux-amd64.tar  20%  15.05M  1.20MB/s    eta 55s           x-amd64.tar.  20%  15.36M  1.20MB/s    eta 55s       -amd64.tar.g  20%  15.67M  1.20MB/s    eta 55s   amd64.tar.gz  21%  15.87M  1.19MB/s    eta 55s                                                                      md64.tar.gz   21%  16.19M  1.06MB/s    eta 55s                                                                      d64.tar.gz    23%  17.33M  1.30MB/s    eta 55s                                                                      64.tar.gz     23%  17.62M  1.30MB/s    eta 55s                                                                      4.tar.gz      23%  17.91M  1.33MB/s    eta 55s                                                                      .tar.gz       24%  18.19M  1.39MB/s    eta 55s                                                                      tar.gz        24%  18.47M  1.40MB/s    eta 49s                                                                      ar.gz         25%  18.72M  1.42MB/s    eta 49s                                                                      r.gz          25%  19.03M  1.43MB/s    eta 49s                                                                      .gz           25%  19.30M  1.41MB/s    eta 49s                                                                      gz            26%  19.56M  1.42MB/s    eta 49s                                                                      z             26%  19.87M  1.43MB/s    eta 47s                                                                                    27%  20.22M  1.43MB/s    eta 47s                                                                                 g  27%  20.53M  1.42MB/s    eta 47s                                                                             go  27%  20.84M  1.41MB/s    eta 47s                                                                        go1  27%  20.92M  1.41MB/s    eta 47s                                                                   go1.  28%  20.97M  1.28MB/s    eta 47s                                                              go1.2  29%  21.80M  1.53MB/s    eta 47s                                                         go1.24  30%  22.48M  1.40MB/s    eta 47s                                                    go1.24.  30%  22.66M  1.37MB/s    eta 44s                                               go1.24.0  30%  22.95M  1.38MB/s    eta 44s                                          go1.24.0.  31%  23.22M  1.37MB/s    eta 44s                                     go1.24.0.l  31%  23.48M  1.37MB/s    eta 44s                                go1.24.0.li  31%  23.76M  1.37MB/s    eta 44s                           go1.24.0.lin  32%  24.05M  1.37MB/s    eta 43s                       o1.24.0.linu  32%  24.31M  1.37MB/s    eta 43s                   1.24.0.linux  32%  24.53M  1.25MB/s    eta 43s               .24.0.linux-  33%  25.22M  1.35MB/s    eta 43s           24.0.linux-a  34%  25.55M  1.35MB/s    eta 41s       4.0.linux-am  34%  25.84M  1.35MB/s    eta 41s   .0.linux-amd  35%  26.17M  1.34MB/s    eta 41s                                                                      0.linux-amd6  35%  26.34M  1.27MB/s    eta 41s                                                                      .linux-amd64  36%  27.42M  1.56MB/s    eta 41s                                                                      linux-amd64.  37%  27.73M  1.45MB/s    eta 41s                                                                      inux-amd64.t  37%  28.00M  1.35MB/s    eta 41s                                                                      nux-amd64.ta  37%  28.26M  1.37MB/s    eta 41s                                                                      ux-amd64.tar  38%  28.58M  1.37MB/s    eta 38s                                                                      x-amd64.tar.  38%  28.83M  1.37MB/s    eta 38s                                                                      -amd64.tar.g  38%  29.14M  1.38MB/s    eta 38s                                                                      amd64.tar.gz  39%  29.44M  1.38MB/s    eta 38s                                                                      md64.tar.gz   39%  29.69M  1.37MB/s    eta 38s                                                                      d64.tar.gz    40%  29.97M  1.37MB/s    eta 37s                                                                      64.tar.gz     40%  30.30M  1.51MB/s    eta 37s                                                                      4.tar.gz      40%  30.64M  1.41MB/s    eta 37s                                                                   .tar.gz       41%  30.94M  1.41MB/s    eta 37s                                                               tar.gz        41%  31.16M  1.31MB/s    eta 36s                                                           ar.gz         42%  31.59M  1.54MB/s    eta 36s                                                       r.gz          43%  32.16M  1.39MB/s    eta 36s                                                   .gz           43%  32.39M  1.38MB/s    eta 36s                                               gz            43%  32.69M  1.39MB/s    eta 36s                                           z             44%  32.95M  1.38MB/s    eta 34s                                                     44%  33.20M  1.38MB/s    eta 34s                                              g  44%  33.48M  1.38MB/s    eta 34s                                         go  45%  33.69M  1.35MB/s    eta 34s                                    go1  45%  33.98M  1.35MB/s    eta 34s                               go1.  45%  34.25M  1.36MB/s    eta 33s                          go1.2  46%  34.56M  1.36MB/s    eta 33s                     go1.24  46%  34.84M  1.36MB/s    eta 33s                go1.24.  46%  35.12M  1.35MB/s    eta 33s           go1.24.0  47%  35.41M  1.33MB/s    eta 33s      go1.24.0.  47%  35.76M  1.34MB/s    eta 31s                                                                        go1.24.0.l  48%  36.12M  1.45MB/s    eta 31s                                                                       go1.24.0.li  48%  36.42M  1.25MB/s    eta 31s                                                                      go1.24.0.lin  49%  36.83M  1.29MB/s    eta 31s                                                                      o1.24.0.linu  49%  37.33M  1.36MB/s    eta 30s                                                                      1.24.0.linux  50%  37.55M  1.34MB/s    eta 30s                                                                      .24.0.linux-  50%  37.80M  1.34MB/s    eta 30s                                                                      24.0.linux-a  50%  38.05M  1.33MB/s    eta 30s                                                                      4.0.linux-am  51%  38.23M  1.32MB/s    eta 30s                                                                      .0.linux-amd  51%  38.50M  1.32MB/s    eta 29s                                                                      0.linux-amd6  51%  38.72M  1.31MB/s    eta 29s                                                                      .linux-amd64  52%  38.94M  1.29MB/s    eta 29s                                                                      linux-amd64.  52%  39.22M  1.26MB/s    eta 29s                                                                      inux-amd64.t  52%  39.47M  1.26MB/s    eta 29s                                                                   nux-amd64.ta  53%  39.75M  1.26MB/s    eta 28s                                                               ux-amd64.tar  53%  40.00M  1.25MB/s    eta 28s                                                           x-amd64.tar.  53%  40.25M  1.21MB/s    eta 28s                                                       -amd64.tar.g  54%  40.56M  1.21MB/s    eta 28s                                                   amd64.tar.gz  54%  40.83M  1.31MB/s    eta 28s                                               md64.tar.gz   54%  41.11M  1.22MB/s    eta 27s                                           d64.tar.gz    55%  41.44M  1.23MB/s    eta 27s                                       64.tar.gz     55%  41.69M  1.24MB/s    eta 27s                                   4.tar.gz      56%  41.97M  1.26MB/s    eta 27s                               .tar.gz       56%  42.34M  1.29MB/s    eta 27s                           tar.gz        57%  42.64M  1.21MB/s    eta 26s                       ar.gz         57%  43.03M  1.26MB/s    eta 26s                   r.gz          58%  43.55M  1.34MB/s    eta 26s               .gz           58%  43.80M  1.35MB/s    eta 26s           gz            58%  44.08M  1.36MB/s    eta 26s       z             59%  44.31M  1.35MB/s    eta 24s   go1.24.0.linux-amd64.tar.gz  100%[==============================================>]  74.75M  1.23MB/s    in 64s     

2026-02-16 15:47:10 (1.16 MB/s) - ‘go1.24.0.linux-amd64.tar.gz’ saved [78382844/78382844]

faris@faris:~/Downloads/toughradius/toughradius$         tar -C $HOME/go -xzf go1.24.0.linux-amd64.tar.gz;
faris@faris:~/Downloads/toughradius/toughradius$         export PATH=$HOME/go/go/bin:$PATH;
faris@faris:~/Downloads/toughradius/toughradius$         go version;
go version go1.24.0 linux/amd64
faris@faris:~/Downloads/toughradius/toughradius$     
faris@faris:~/Downloads/toughradius/toughradius$ ^C
faris@faris:~/Downloads/toughradius/toughradius$ export PATH=$HOME/go/go/bin:$PATH; go run main.go -initdb -c toughradius.dev.yml
go: downloading golang.org/x/sync v0.17.0
go: downloading gopkg.in/yaml.v3 v3.0.1
go: downloading github.com/bwmarrin/snowflake v0.3.0
go: downloading github.com/pkg/errors v0.9.1
go: downloading github.com/360EntSecGroup-Skylar/excelize v1.4.1
go: downloading github.com/gocarina/gocsv v0.0.0-20230616125104-99d496ca653d
go: downloading github.com/golang-jwt/jwt/v4 v4.5.2
go: downloading github.com/labstack/echo-jwt/v4 v4.3.1
go: downloading github.com/labstack/echo/v4 v4.13.4
go: downloading github.com/labstack/gommon v0.4.2
go: downloading github.com/spf13/cast v1.10.0
go: downloading go.uber.org/zap v1.27.0
go: downloading github.com/panjf2000/ants/v2 v2.11.3
go: downloading gorm.io/gorm v1.31.1
go: downloading layeh.com/radius v0.0.0-20231213012653-1006025d24f8
go: downloading github.com/glebarez/sqlite v1.11.0
go: downloading github.com/robfig/cron/v3 v3.0.1
go: downloading github.com/shirou/gopsutil/v4 v4.25.10
go: downloading gopkg.in/natefinch/lumberjack.v2 v2.2.1
go: downloading gorm.io/driver/postgres v1.6.0
go: downloading github.com/golang-jwt/jwt/v5 v5.3.0
go: downloading github.com/stretchr/testify v1.11.1
go: downloading github.com/go-playground/validator/v10 v10.28.0
go: downloading github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de
go: downloading github.com/mattn/go-colorable v0.1.14
go: downloading github.com/mattn/go-isatty v0.0.20
go: downloading github.com/valyala/fasttemplate v1.2.2
go: downloading github.com/jinzhu/now v1.1.5
go: downloading go.uber.org/multierr v1.11.0
go: downloading github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
go: downloading golang.org/x/crypto v0.43.0
go: downloading golang.org/x/net v0.46.0
go: downloading golang.org/x/time v0.14.0
go: downloading github.com/jinzhu/inflection v1.0.0
go: downloading golang.org/x/text v0.30.0
go: downloading github.com/tklauser/go-sysconf v0.3.15
go: downloading golang.org/x/sys v0.37.0
go: downloading github.com/jackc/pgx/v5 v5.7.6
go: downloading github.com/valyala/bytebufferpool v1.0.0
go: downloading github.com/gabriel-vasile/mimetype v1.4.10
go: downloading github.com/go-playground/universal-translator v0.18.1
go: downloading github.com/leodido/go-urn v1.4.0
go: downloading github.com/glebarez/go-sqlite v1.22.0
go: downloading modernc.org/sqlite v1.40.1
go: downloading github.com/davecgh/go-spew v1.1.1
go: downloading github.com/pmezard/go-difflib v1.0.0
go: downloading github.com/tklauser/numcpus v0.10.0
go: downloading github.com/jackc/pgpassfile v1.0.0
go: downloading github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
go: downloading github.com/jackc/puddle/v2 v2.2.2
go: downloading github.com/go-playground/locales v0.14.1
go: downloading modernc.org/libc v1.67.1
go: downloading golang.org/x/exp v0.0.0-20251023183803-a4bb9ffd2546
go: downloading github.com/dustin/go-humanize v1.0.1
go: downloading github.com/google/uuid v1.6.0
go: downloading modernc.org/mathutil v1.7.1
go: downloading modernc.org/memory v1.11.0
go: downloading github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec
web/static.go:11:12: pattern dist/*: no matching files found
faris@faris:~/Downloads/toughradius/toughradius$ cd web && npm install && npm run build

up to date, audited 319 packages in 4s

70 packages are looking for funding
  run `npm fund` for details

5 vulnerabilities (4 moderate, 1 high)

To address issues that do not require attention, run:
  npm audit fix

To address all issues (including breaking changes), run:
  npm audit fix --force

Run `npm audit` for details.

> toughradius-web@9.0.0 build
> tsc && vite build

vite v5.4.21 building for production...
✓ 13872 modules transformed.
dist/admin/index.html                           0.68 kB │ gzip:   0.39 kB
dist/admin/assets/react-vendor-Cwh1aMWO.js    141.41 kB │ gzip:  45.48 kB
dist/admin/assets/index-CbjDT_xX.js           265.22 kB │ gzip:  54.31 kB
dist/admin/assets/react-admin-C8Lb0puc.js   1,013.17 kB │ gzip: 293.92 kB
dist/admin/assets/echarts-By_BdcVi.js       1,052.10 kB │ gzip: 349.70 kB

(!) Some chunks are larger than 500 kB after minification. Consider:
- Using dynamic import() to code-split the application
- Use build.rollupOptions.output.manualChunks to improve chunking: https://rollupjs.org/configuration-options/#output-manualchunks
- Adjust chunk size limit for this warning via build.chunkSizeWarningLimit.
✓ built in 13.32s
faris@faris:~/Downloads/toughradius/toughradius/web$ cd /home/faris/Downloads/toughradius/toughradius
faris@faris:~/Downloads/toughradius/toughradius$ export PATH=$HOME/go/go/bin:$PATH; go run main.go -initdb -c toughradius.dev.yml
2026-02-16T20:50:58.113+0800    INFO    app/database.go:24      SQLite database path: rundata/data/toughradius.db
2026-02-16T20:50:58.114+0800    INFO    toughradius/main.go:77  Database connection successful, type: sqlite
2026-02-16T20:50:58.175+0800    INFO    app/config_manager.go:93        config schemas loaded from JSON {"count": 9}
2026-02-16T20:50:58.175+0800    INFO    app/config_manager.go:85        config loaded from database     {"count": 0}
faris@faris:~/Downloads/toughradius/toughradius$ export PATH=$HOME/go/go/bin:$PATH; go run main.go -c toughradius.dev.yml
2026-02-16T20:51:14.917+0800    INFO    app/database.go:24      SQLite database path: rundata/data/toughradius.db
2026-02-16T20:51:14.918+0800    INFO    toughradius/main.go:77  Database connection successful, type: sqlite
2026-02-16T20:51:14.925+0800    INFO    app/config_manager.go:93        config schemas loaded from JSON {"count": 9}
2026-02-16T20:51:14.925+0800    INFO    app/config_manager.go:85        config loaded from database     {"count": 0}
2026-02-16T20:51:14.925+0800    INFO    toughradius/main.go:115 Starting Radius Resec server on 0.0.0.0:2083
2026-02-16T20:51:14.925+0800    ERROR   toughradius/main.go:115 Radius Resec server error: read rundata: is a directory
main.main.func4
        /home/faris/Downloads/toughradius/toughradius/main.go:115
golang.org/x/sync/errgroup.(*Group).Go.func1
        /home/faris/go/pkg/mod/golang.org/x/sync@v0.17.0/errgroup/errgroup.go:93
2026-02-16T20:51:14.925+0800    INFO    toughradius/main.go:101 Starting Radius Auth server on 0.0.0.0:1812
2026-02-16T20:51:14.925+0800    INFO    toughradius/main.go:106 Starting Radius Acct server on 0.0.0.0:1813
2026-02-16T20:51:14.925+0800    INFO    webserver/server.go:87  React Admin static files loaded successfully
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/auth.go:28     Add API POST Router /api/v1/auth/login
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/auth.go:29     Add API GET Router /api/v1/auth/me
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/users.go:247   Add API GET Router /api/v1/users
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/users.go:248   Add API GET Router /api/v1/users/:id
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/users.go:249   Add API POST Router /api/v1/users
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/users.go:250   Add API PUT Router /api/v1/users/:id
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/users.go:251   Add API DELETE Router /api/v1/users/:id
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/dashboard.go:112       Add API GET Router /api/v1/dashboard/stats
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/profiles.go:431        Add API GET Router /api/v1/radius-profiles
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/profiles.go:432        Add API GET Router /api/v1/radius-profiles/:id
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/profiles.go:433        Add API POST Router /api/v1/radius-profiles
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/profiles.go:434        Add API PUT Router /api/v1/radius-profiles/:id
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/profiles.go:435        Add API DELETE Router /api/v1/radius-profiles/:id
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/accounting.go:164      Add API GET Router /api/v1/accounting
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/accounting.go:165      Add API GET Router /api/v1/accounting/:id
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/sessions.go:254        Add API GET Router /api/v1/sessions
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/sessions.go:255        Add API GET Router /api/v1/sessions/:id
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/sessions.go:256        Add API DELETE Router /api/v1/sessions/:id
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/nas.go:299     Add API GET Router /api/v1/network/nas
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/nas.go:300     Add API GET Router /api/v1/network/nas/:id
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/nas.go:301     Add API POST Router /api/v1/network/nas
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/nas.go:302     Add API PUT Router /api/v1/network/nas/:id
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/nas.go:303     Add API DELETE Router /api/v1/network/nas/:id
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/settings.go:30 Add API GET Router /api/v1/system/settings
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/settings.go:31 Add API GET Router /api/v1/system/settings/:id
2026-02-16T20:51:14.925+0800    DEBUG   adminapi/settings.go:32 Add API GET Router /api/v1/system/config/schemas
2026-02-16T20:51:14.926+0800    DEBUG   adminapi/settings.go:33 Add API POST Router /api/v1/system/settings
2026-02-16T20:51:14.926+0800    DEBUG   adminapi/settings.go:34 Add API PUT Router /api/v1/system/settings/:id
2026-02-16T20:51:14.926+0800    DEBUG   adminapi/settings.go:35 Add API DELETE Router /api/v1/system/settings/:id
2026-02-16T20:51:14.926+0800    DEBUG   adminapi/settings.go:36 Add API POST Router /api/v1/system/config/reload
2026-02-16T20:51:14.926+0800    DEBUG   adminapi/nodes.go:31    Add API GET Router /api/v1/network/nodes
2026-02-16T20:51:14.926+0800    DEBUG   adminapi/nodes.go:32    Add API GET Router /api/v1/network/nodes/:id
2026-02-16T20:51:14.926+0800    DEBUG   adminapi/nodes.go:33    Add API POST Router /api/v1/network/nodes
2026-02-16T20:51:14.926+0800    DEBUG   adminapi/nodes.go:34    Add API PUT Router /api/v1/network/nodes/:id
2026-02-16T20:51:14.926+0800    DEBUG   adminapi/nodes.go:35    Add API DELETE Router /api/v1/network/nodes/:id
2026-02-16T20:51:14.926+0800    DEBUG   adminapi/operators.go:33        Add API GET Router /api/v1/system/operators/me
2026-02-16T20:51:14.926+0800    DEBUG   adminapi/operators.go:34        Add API PUT Router /api/v1/system/operators/me
2026-02-16T20:51:14.926+0800    DEBUG   adminapi/operators.go:37        Add API GET Router /api/v1/system/operators
2026-02-16T20:51:14.926+0800    DEBUG   adminapi/operators.go:38        Add API GET Router /api/v1/system/operators/:id
2026-02-16T20:51:14.926+0800    DEBUG   adminapi/operators.go:39        Add API POST Router /api/v1/system/operators
2026-02-16T20:51:14.926+0800    DEBUG   adminapi/operators.go:40        Add API PUT Router /api/v1/system/operators/:id
2026-02-16T20:51:14.926+0800    DEBUG   adminapi/operators.go:41        Add API DELETE Router /api/v1/system/operators/:id
2026-02-16T20:51:14.926+0800    INFO    webserver/server.go:57  Start the management server 0.0.0.0:1816
2026-02-16T20:51:14.926+0800    INFO    runtime/asm_amd64.s:1700        Prepare to start the TLS management port 0.0.0.0:0
⇨ http server started on [::]:1816
2026-02-16T20:51:14.926+0800    ERROR   runtime/asm_amd64.s:1700        Error starting TLS management port open rundata/private/toughradius.tls.crt: permission denied
runtime.goexit
        /home/faris/go/go/src/runtime/asm_amd64.s:1700
2026-02-16T20:51:17.930+0800    INFO    app/app.go:134  initialized default super admin account {"username": "admin"}
2026-02-16T20:51:17.934+0800    INFO    app/app.go:135  initialized config      {"key": "radius.EapMethod", "default": "eap-md5"}
2026-02-16T20:51:17.937+0800    INFO    app/app.go:135  initialized config      {"key": "radius.EapEnabledHandlers", "default": "*"}
2026-02-16T20:51:17.941+0800    INFO    app/app.go:135  initialized config      {"key": "radius.IgnorePassword", "default": "false"}
2026-02-16T20:51:17.944+0800    INFO    app/app.go:135  initialized config      {"key": "radius.AccountingHistoryDays", "default": "90"}
2026-02-16T20:51:17.946+0800    INFO    app/app.go:135  initialized config      {"key": "radius.AcctInterimInterval", "default": "300"}
2026-02-16T20:51:17.948+0800    INFO    app/app.go:135  initialized config      {"key": "radius.SessionTimeout", "default": "3600"}
2026-02-16T20:51:17.950+0800    INFO    app/app.go:135  initialized config      {"key": "radius.LogLevel", "default": "info"}
2026-02-16T20:51:17.953+0800    INFO    app/app.go:135  initialized config      {"key": "radius.RejectDelayMaxRejects", "default": "7"}
2026-02-16T20:51:17.955+0800    INFO    app/app.go:135  initialized config      {"key": "radius.RejectDelayWindowSeconds", "default": "10"}
faris@faris:~/Downloads/toughradius/toughradius$ 
faris@faris:~/Downloads/toughradius/toughradius$ ^C
faris@faris:~/Downloads/toughradius/toughradius$ ls -lt rundata/logs
total 0
faris@faris:~/Downloads/toughradius/toughradius$ curl -I http://localhost:1816
HTTP/1.1 405 Method Not Allowed
Allow: OPTIONS, GET
Vary: Accept-Encoding
Date: Mon, 16 Feb 2026 12:52:39 GMT

faris@faris:~/Downloads/toughradius/toughradius$ 

--------------
## Auto-Deploy to coolify
Updated workflow with auto-deploy to Coolify.

**To enable auto-deploy:**

1. **Get Coolify Webhook URL:**
   - Go to your Coolify dashboard
   - Open your Radio project
   - Click on **Deploy** tab
   - Find **Webhook** section
   - Copy the webhook URL

2. **Add to GitHub Secrets:**
   - Go to your GitHub repository → Settings → Secrets → Actions
   - Add new secret:
     - **Name**: `COOLIFY_WEBHOOK_URL`
     - **Secret**: `https://your-coolify-domain/api/deploy/some-id`

3. **Push to trigger deployment:**
   ```bash
   git add .
   git commit -m "feat: add auto-deploy to coolify"
   git push origin main
   ```

After push, GitHub will:
1. Build Docker image
2. Push to Docker Hub (`farisnoaman/toughradius`)
3. Trigger Coolify to pull new image and redeploy

---
## Docker Compose
Updated `docker-compose.yml` for production with PostgreSQL.

**Changes made:**
- Uses your custom image: `farisnoaman/toughradius:latest`
- Added required environment variables (domain, web secret, logging)
- Added `restart: unless-stopped` for both services
- Added `depends_on` to ensure Postgres starts first
- Added proper volume definitions

**To deploy:**
```bash
# Update the password in docker-compose.yml first
# Then run:
docker-compose up -d
```

**Update these values before running:**
- `TOUGHRADIUS_SYSTEM_DOMAIN` - your domain
- `TOUGHRADIUS_WEB_SECRET` - generate a random secret
- `TOUGHRADIUS_DB_PASSWD` - strong password
- `POSTGRES_PASSWORD` - strong password (should match DB password)