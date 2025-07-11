$defFile = "postgres.def"
$outDir  = Resolve-Path '..\output' -ErrorAction SilentlyContinue -ErrorVariable _dummy
if (-not $outDir) { $outDir = (New-Item -ItemType Directory -Path '..\output').FullName }
$outFile = Join-Path $outDir 'postgres.exe'

function TryVS {
    $vswhere = "$Env:ProgramFiles (x86)\Microsoft Visual Studio\Installer\vswhere.exe"
    if (-not (Test-Path $vswhere)) { return $false }
    $vsRoot = & $vswhere -latest `
	                     -products * `
                         -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 `
                         -property installationPath |
              Select-Object -First 1
    if (-not $vsRoot) { return $false }
    $msvcDir = Get-ChildItem -Path (Join-Path $vsRoot 'VC\Tools\MSVC') |
               Sort-Object Name -Descending |
               Select-Object -First 1
    $linkExe = Join-Path $msvcDir.FullName 'bin\Hostx64\x64\link.exe'
    & cmd /c "`"$vsRoot\VC\Auxiliary\Build\vcvars64.bat`" >nul `&`& `"$linkExe`" /DLL /NOENTRY /DEF:$defFile /OUT:`"$outFile`""
    return $true
}

function TryGCC {
    $gcc = (& where.exe gcc.exe 2>$null | Select-Object -First 1)
    if (-not $gcc) { return $false }
    $args = @(
        "-shared",
        "-nostdlib",
        $defFile,
        "-o", $outFile
    )
    & $gcc @args
    return $true
}

function TryClang {
    $lld = (& where.exe lld-link.exe 2>$null | Select-Object -First 1)
    if ($lld) {
        & $lld /DLL /NOENTRY /DEF:$defFile /OUT:"$outFile"
        return $true
    }
    $clang = (& where.exe clang.exe 2>$null | Select-Object -First 1)
    if (-not $clang) { return $false }
    $args = @(
        "-shared",
        "-nostdlib",
        $defFile,
        "-o", $outFile
    )
    & $clang @args
    return $true
}

if (TryVS)    { Write-Host "Definition file built using Visual Studio"; exit 0 }
if (TryGCC)   { Write-Host "Definition file built using GCC"; exit 0 }
if (TryClang) { Write-Host "Definition file built using Clang"; exit 0 }

throw "Could not build the definition file"