# Script untuk clean up Fabric Network sepenuhnya
Write-Host "Cleaning up Fabric Network completely..." -ForegroundColor Red

# Stop all containers
Write-Host "Stopping all containers..." -ForegroundColor Yellow
wsl docker stop $(wsl docker ps -aq)

# Remove all containers
Write-Host "Removing all containers..." -ForegroundColor Yellow
wsl docker rm $(wsl docker ps -aq)

# Remove all volumes
Write-Host "Removing all volumes..." -ForegroundColor Yellow
wsl docker volume prune -f

# Clean up test network
Write-Host "Cleaning up test network..." -ForegroundColor Yellow
wsl bash -c 'cd ~/fabric-workspace/fabric-samples/test-network && ./network.sh down'

# Remove any leftover chaincode images
Write-Host "Removing chaincode images..." -ForegroundColor Yellow
wsl docker rmi $(wsl docker images --filter "reference=dev-peer*" -q) 2>/dev/null

Write-Host "Complete cleanup finished!" -ForegroundColor Green
Write-Host "Now you can run: .\start-fabric-only.ps1" -ForegroundColor Cyan