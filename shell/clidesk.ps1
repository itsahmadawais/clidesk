# CLIDesk PowerShell wrapper
# Launches clidesk and cd's to whatever directory you ended up in.
#
# SETUP — add this to your PowerShell profile ($PROFILE):
#
#   . "C:\path\to\clidesk\shell\clidesk.ps1"
#
# Then just type:  clidesk
# Or with a theme: clidesk --theme dracula

function clidesk {
    $tmp = [System.IO.Path]::GetTempFileName()
    try {
        & clidesk.exe @args --print-dir $tmp
        if (Test-Path $tmp) {
            $dir = Get-Content $tmp -Raw
            if ($dir -and (Test-Path $dir -PathType Container)) {
                Set-Location $dir
            }
        }
    } finally {
        Remove-Item $tmp -ErrorAction SilentlyContinue
    }
}
