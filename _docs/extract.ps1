param(
    [string]$Src1, [string]$Src2, [string]$Src3, [string]$OutDir
)
$ErrorActionPreference = 'Stop'
$out = $OutDir
New-Item -ItemType Directory -Force -Path $out | Out-Null

function Extract-Docx($src, $dst) {
    Write-Host ("SRC=[" + $src + "]  DST=[" + $dst + "]")
    $tmp = Join-Path $env:TEMP ('docx_' + [Guid]::NewGuid().ToString('N'))
    $zipCopy = $tmp + '.zip'
    Copy-Item -LiteralPath $src -Destination $zipCopy -Force
    Expand-Archive -LiteralPath $zipCopy -DestinationPath $tmp -Force
    Remove-Item -LiteralPath $zipCopy -Force
    $xml = Get-Content -LiteralPath (Join-Path $tmp 'word\document.xml') -Raw -Encoding UTF8
    $xml = $xml -replace '</w:p>', "`r`n"
    $xml = $xml -replace '<w:tab[^/]*/>', "`t"
    $xml = $xml -replace '<w:br[^/]*/>', "`r`n"
    $xml = $xml -replace '<[^>]+>', ''
    $xml = [System.Net.WebUtility]::HtmlDecode($xml)
    Set-Content -LiteralPath $dst -Value $xml -Encoding UTF8
    Remove-Item -Recurse -Force $tmp
    $size = (Get-Item $dst).Length
    Write-Host "Wrote $dst ($size bytes)"
}

Extract-Docx $Src1 (Join-Path $out 'directory.txt')
Extract-Docx $Src2 (Join-Path $out 'solution.txt')

# PDF: try pdftotext if available, else fallback to Word COM
$pdf = $Src3
$pdfOut = Join-Path $out 'guide.txt'
$pdftotext = Get-Command pdftotext -ErrorAction SilentlyContinue
if ($pdftotext) {
    & pdftotext.exe -layout -enc UTF-8 $pdf $pdfOut
    Write-Host "pdftotext ok"
} else {
    Write-Host "pdftotext not found; trying Word COM"
    try {
        $word = New-Object -ComObject Word.Application
        $word.Visible = $false
        $doc = $word.Documents.Open($pdf, $false, $true)
        $doc.SaveAs([ref]$pdfOut, [ref]7)  # 7 = wdFormatText
        $doc.Close($false)
        $word.Quit()
        Write-Host "Word COM ok"
    } catch {
        Write-Host ("Word COM failed: " + $_.Exception.Message)
    }
}

Get-ChildItem $out -File | Format-Table Name, Length
