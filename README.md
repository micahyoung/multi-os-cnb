# Multi-OS buildpack demo

## Requirements
* Docker Desktop
* Pack 0.5.1 or greater

## Usage
### Windows
```powershell
pack package-buildpack multi-os-cnb:windows --config package-windows.yml

pack build multi-os-test:windows `
  --buildpack docker://multi-os-cnb:windows `
  --builder cnbs/sample-builder:nanoserver-1809 `
  --path integration/testdata/app 
  
docker run -i --rm multi-os-test:windows
```

### Linux
```
pack package-buildpack multi-os-cnb:linux --config package-linux.yml

pack build multi-os-test:linux \
  --buildpack docker://multi-os-cnb:linux \
  --builder cnbs/sample-builder:bionic \
  --path integration/testdata/app 
  
docker run -i --rm multi-os-test:linux
```
