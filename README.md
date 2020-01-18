# golang-http-file-upload-download

 A simple ready-made golang server that enables you to uploads files

 Start with

 ```
 go run main.go
 ```

uses unix ```.``` to hide files until they have finished uploading

## Configuration

Set environment variables

- ```GOLANGHTTPUPLOAD_UPLOAD_PATH```  
  directory where files should be uploaded to
- ```GOLANGHTTPUPLOAD_SERVE_PATH```  
  directory where html files are served from
- ```GOLANGHTTPUPLOAD_BINDIP_PORT```  
  ip and port to bind to. Default is ```:8080```

