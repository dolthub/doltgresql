$ErrorActionPreference = "Stop"

$vswhere = "$Env:ProgramFiles (x86)\Microsoft Visual Studio\Installer\vswhere.exe"
if (-not (Test-Path $vswhere)) {
    throw "vswhere.exe not found at expected location: $vswhere"
}

$vsRoot = & $vswhere -latest -products * `
                     -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 `
                     -property installationPath |
          Select-Object -First 1

if (-not $vsRoot) { throw "No suitable Visual Studio installation found." }

$msvcDir = Get-ChildItem -Path (Join-Path $vsRoot 'VC\Tools\MSVC') |
           Sort-Object Name -Descending |
           Select-Object -First 1
$linkExe = Join-Path $msvcDir.FullName 'bin\Hostx64\x64\link.exe'

$outDir  = Resolve-Path '..\output' -ErrorAction SilentlyContinue `
           -ErrorVariable _dummy
if (-not $outDir) { $outDir = (New-Item -ItemType Directory -Path '..\output').FullName }

& cmd /c "`"$vsRoot\VC\Auxiliary\Build\vcvars64.bat`" >nul `&`& `"$linkExe`" /DLL /NOENTRY /DEF:postgres.def /OUT:`"$outDir\postgres.exe`""
