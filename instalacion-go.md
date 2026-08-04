# Instalación de Go.



1.Descargar el binario de Go desde su sitio oficial
Ve a go.dev/dl para copiar el enlace de la versión más reciente para Linux. Descárgalo usando wget:
```bash
# Reemplaza el enlace con la versión actual
wget https://go.dev/dl/go1.X.X.linux-amd64.tar.gz 
``` 

en mi caso fue:  https://go.dev/dl/go1.26.4.linux-amd64.tar.gz

2.Limpiar y extraer:
(Requiere permisos sudo)

Elimina cualquier instalación antigua de Go para evitar conflictos (si no tienes ninguna, el comando no hará daño) y extrae el nuevo archivo en /usr/local:

```bash
# Reemplaza el enlace con la versión actual
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.26.4.linux-amd64.tar.gz
``` 

3.Añadir al PATH en Zsh:Configuración de variables de entorno.Para que el sistema reconozca el comando go, debes agregarlo a tu ruta. Como utilizas Zsh, inyecta la variable en tu archivo de configuración y recárgalo
Bash
``` 
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc
source ~/.zshrc
``` 

4. Verificar la instalación: Asegúrate de que todo está configurado correctamente revisando la versión instalada:

``` Bash
go version
``` 

Alternativa rápida: Si prefieres gestionar tus paquetes globales usando Snap en lugar de hacerlo manualmente, puedes instalarlo con un solo comando: sudo snap install go --classic.
