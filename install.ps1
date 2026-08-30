$link = "https://github.com/Narehatechi/Narecord/releases/latest/download/NarelotlCli.exe"

$outfile = "$env:TEMP\NarelotlCli.exe"

Write-Output "Downloading installer to $outfile"

Invoke-WebRequest -Uri "$link" -OutFile "$outfile"

Write-Output ""

Start-Process -Wait -NoNewWindow -FilePath "$outfile"

# Cleanup
Remove-Item -Force "$outfile"
