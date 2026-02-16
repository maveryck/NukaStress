 <img width="1272" height="805" alt="image" src="https://github.com/user-attachments/assets/25ae511e-ac52-4040-a025-232ad98e1dae" />

 # NukaStress


Sobrevive al apocalipsis nuclear de tu hardware.

Slogan oficial:
"Sobrevive al apocalipsis nuclear de tu hardware - Prueba tu PC como un experimento Vault-Tec!"

NukaStress es una suite de stress testing en Go con interfaz desktop (Fyne) estilo Pip-Boy.
Permite ejecutar pruebas extremas de CPU, GPU, RAM, disco, red y chequeos de telemetria, con
mecanismos de seguridad para abortar automaticamente cuando el sistema entra en zona critica.

## Estado actual (MVP funcional)

- GUI Fallout/Pip-Boy con 4 tabs:
  - `Wasteland Tests`
  - `Pip-Boy Monitor`
  - `Nuka-Config`
  - `Rad-Reports`
- Motor central con ejecucion secuencial de pruebas y acumulacion de errores
- Monitoreo en vivo (CPU, RAM, temperatura)
- Guardas de seguridad con hard-stop por:
  - limite de temperatura
  - exceso de errores
- Boton de panico: `Evacuar Vault`
- Reporte HTML portable generado en carpeta temporal del sistema
- Modo headless para servidor/CI

## Arquitectura del proyecto

```text
NukaStress/
  main.go                 # Entrada GUI/headless
  core/
    engine.go             # Orquestador principal (config, ejecucion, seguridad)
    types.go              # Tipos base (Config, Result, Snapshot, Alert)
  gui/
    main_window.go        # Ventana principal y tabs
    theme.go              # Tema Pip-Boy
  monitor/
    pipboy_sensors.go     # Telemetria en vivo con gopsutil
  tests/
    cpu_nuke.go           # Stress CPU multihilo
    gpu_rad.go            # GPU via memtest_vulkan o fallback
    mem_vault.go          # Stress RAM por patrones aleatorios
    psu_quantum.go        # Chequeo PSU/telemetria
    extras/
      disk_io.go          # Burst de escritura temporal en disco
      net_flood.go        # Probe UDP por intentos
  report/
    html.go               # Generacion de reporte HTML
  utils/
    fallout_logger.go     # Utilidades de logs tematicos
```

## Requisitos

- Go `1.22+`
- Dependencias principales:
  - `fyne.io/fyne/v2`
  - `github.com/shirou/gopsutil/v4`
  - `github.com/mattn/go-colorable`
  - `gonum.org/v1/plot`
- Opcional para ruta GPU real:
  - `memtest_vulkan` en `PATH`

## Ejecucion rapida (GUI)

```bash
go mod tidy
go run .
```

La app abre una ventana `NukaStress - Sobrevive al apocalipsis nuclear` y activa telemetria en segundo plano.

## Modo Headless

```bash
go run . --headless --minutes 10
```

Flags disponibles:

- `--headless`: ejecuta sin UI
- `--minutes N`: duracion total configurada para la corrida (default en `main.go`: 10)

Salida esperada: resumen por prueba en consola con `PASS/FAIL`, errores y mensaje.

## Flujo interno del motor (`core.Engine`)

1. Carga configuracion por defecto:
   - `Duration: 5m`
   - `TargetHost: 1.1.1.1:53`
   - `MaxTempC: 92`
   - `MaxErrorCount: 5`
   - `GPUBackend: auto`
   - `Mode: beginner`
2. Inicia stream de telemetria (1 muestra/segundo)
3. Ejecuta pruebas secuenciales en este orden:
   - CPU
   - GPU
   - RAM
   - PSU
   - Disco
   - Red
4. Acumula errores; si supera `MaxErrorCount`, aborta y agrega resultado `SafetyGuard`
5. Si durante ejecucion la temperatura supera `MaxTempC`, dispara alerta critica y cancela corrida

## Modos de operacion en GUI

Selector en `Wasteland Tests`:

- `Beginner`
  - `ModeBeginner`
  - `MaxTempC = 92`
  - `MaxErrorCount = 5`
- `Wasteland`
  - `ModeWasteland`
  - `MaxTempC = 95`
  - `MaxErrorCount = 20`

La duracion se ajusta con slider (`1-60` minutos).

## Detalle de pruebas

### CPU (`tests.StressCPUNuke`)

- Usa todos los hilos (`runtime.NumCPU`) si no se especifican
- Perfila vendor (`intel-avx`, `amd-fma`, `unknown`) de forma informativa
- Ejecuta carga matematica continua y cuenta errores numericos (NaN/Inf)

### GPU (`tests.StressGPURad`)

- Si encuentra `memtest_vulkan`:
  - lo ejecuta con timeout
  - marca FAIL si detecta patrones de error o falla el comando
- Si no existe:
  - usa fallback de carga temporizada

### RAM (`tests.StressMemVault`)

- Buffer de 16MB
- Escritura/verificacion aleatoria de bytes
- Marca corrupcion si detecta inconsistencias

### PSU (`tests.StressPSUQuantum`)

- Muestreo de sensores de temperatura (gopsutil)
- Verificacion basica de estabilidad de telemetria

### Disco (`extras.DiskBurst`)

- Escribe archivo temporal en `os.TempDir()` (~64MB)
- Fuerza `Sync` y elimina archivo al finalizar

### Red (`extras.NetProbe`)

- Dial UDP a `TargetHost` por intentos (default 20)
- Cuenta fallos de conectividad

## Monitoreo y alertas

NukaStress almacena historial de snapshots (`max 300`) con:

- `% CPU`
- `% RAM`
- `Temperatura promedio`
- `Timestamp`

Alertas criticas actuales:

- `TEMP_LIMIT`: temperatura por encima del umbral
- `ERROR_LIMIT`: exceso de errores acumulados

## Reportes

Boton `Generar Rad-Report` (GUI):

- crea archivo HTML en carpeta temporal del sistema
- nombre: `nukastress_report_YYYYMMDD_HHMMSS.html`
- incluye:
  - tabla de resultados
  - estado final de telemetria (ultima muestra)

## Build Windows portable

```bat
build_windows.bat
```

El script genera:

- `NukaStress.exe` (portable, `windowsgui`, optimizado con `-s -w`)

## Seguridad y limitaciones actuales

- El stress de PSU hoy es telemetrico, no un modelo electrico de carga real
- El fallback de GPU es sintetico si `memtest_vulkan` no esta instalado
- Sensores dependen de soporte del SO/drivers; algunas temperaturas pueden no estar disponibles
- No hay aun persistencia historica de sesiones fuera del reporte puntual

## Roadmap sugerido

- Perfiles predefinidos por escenario (`Gaming`, `Workstation`, `Server Burn-In`)
- Export JSON/CSV ademas de HTML
- Suite de tests paralela configurable por riesgo
- Integracion de firmas de inestabilidad (ML/rule engine)
- Dashboard de tendencias por sesiones
- Plugin backends para herramientas externas (GPU/IO/network)

## Desarrollo

Comandos utiles:

```bash
go mod tidy
go run .
go run . --headless --minutes 3
go test ./...
```

## Licencia

Define aqui la licencia de distribucion del proyecto (MIT/Apache-2.0/GPL, etc.).
## Releases (GitHub)

Nombre de binarios de release: `Nukastrest`.

Entrypoints:

- GUI: `./cmd/nukastrest-gui`
- CLI: `./cmd/nukastrest-cli`

### Build local rapido

```powershell
./build_release.ps1
```

Salida en `dist/`:

- `Nukastrest-windows-amd64-gui.exe`
- `Nukastrest-windows-amd64-cli.exe`
- `Nukastrest-linux-amd64-cli`

### Build y publish automatico

Workflow: `.github/workflows/release.yml`

- `workflow_dispatch`: compila y sube artifacts.
- `push tag v*` (ejemplo `v1.0.0`): compila y publica GitHub Release automaticamente.

## Wails UI (HTML/CSS/JS)

Se agrego una version desktop con `Wails` y frontend web inspirado en tu referencia Pip-Boy.

Entrypoint Wails:

- `./cmd/nukastrest-wails`

Frontend embebido para Wails:

- `cmd/nukastrest-wails/frontend/dist/index.html`
- `cmd/nukastrest-wails/frontend/dist/style.css`
- `cmd/nukastrest-wails/frontend/dist/app.js`

Build rapido Wails (Windows):

```powershell
./build_wails.ps1
```

Salida:

- `dist/Nukastrest-windows-amd64-wails.exe`

## Linux CLI (Uso Para Usuarios)

No necesitas instalar Go para usar la version CLI en Linux.

1. Descarga el binario `Nukastrest-linux-amd64-cli` desde Releases.
2. Dale permisos de ejecucion:

```bash
chmod +x Nukastrest-linux-amd64-cli
```

3. Ejecuta una prueba rapida (por ejemplo 3 minutos):

```bash
./Nukastrest-linux-amd64-cli --minutes 3
```

4. Ejecucion por defecto (10 minutos):

```bash
./Nukastrest-linux-amd64-cli
```

Salida esperada: resumen `PASS/FAIL` por prueba en consola.

## Linux CLI (Build Para Mantenedores)

Desde Windows, Linux o macOS puedes compilar el binario Linux CLI con cross-compile:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o dist/Nukastrest-linux-amd64-cli ./cmd/nukastrest-cli
```

En este repo tambien puedes usar:

```powershell
./build_release.ps1
```

Ese script deja el archivo en `dist/Nukastrest-linux-amd64-cli`.

## Distribucion recomendada (actualizado)

### Windows (primero)

Descarga y ejecuta directamente el binario principal:

- `Nukastrest-windows-amd64-wails.exe`

### Linux (sin asset en release)

En Linux usamos flujo por clone del repo:

```bash
git clone https://github.com/maveryck/NukaStress.git
cd NukaStress
chmod +x ./dist/Nukastrest-linux-amd64-cli
./dist/Nukastrest-linux-amd64-cli
```
