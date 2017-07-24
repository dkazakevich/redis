# Redis-like in-memory cache

## Installation and upgrade

```
    # Use 'go get' to install or upgrade (-u) the redis package
    go get -u github.com/dkazakevich/redis
    
    # Use `go test` to run tests
    go test github.com/dkazakevich/redis
    
    # Run the redis server
    $GOPATH/bin/redis ($USERPROFILE/go/bin/redis)
    
    # Run the redis server in background
    nohup $GOPATH/bin/redis &
```

By default the server runs on the `8080` port number. Do the following steps to run on another port number:
1. Go to the `$GOPATH/bin/redis` directory
1. Create the `conf.json` file with the following content and indicate required port number:
```
{
  "serverPort": "8081"
}
```

## HTTP Rest Api description

### Put key-value pare into the cache
```
PUT /api/v1/values/{key}
```
 Parameter | Required | Description
-----------|----------|------------------------------------------------------
 key       | yes      | Cache key
 expire    | no       | Set a timeout in seconds on key. After the timeout has expired, the key will automatically be deleted.
 
 _**Examples**_:
```
Request: curl -X PUT -H 'content-type: application/json' -d '"June"' http://localhost:8080/api/v1/values/sixthMonth?expire=20
Response code: 200
Response body: {"value":"June"}

Request: curl -X PUT -H 'content-type: application/json' -d '{"planet1":"Mercury", "planet2":"Venus", "planet3":"Earth"}' http://localhost:8080/api/v1/values/planets
Response code: 200
Response body: {"value":{"planet1":"Mercury","planet2":"Venus","planet3":"Earth"}}
            
Request: curl -X PUT -H 'content-type: application/json' -d '["Toyota","Opel","Ford"]' http://localhost:8080/api/v1/values/cars
Response code: 200
Response body: {"value":["Toyota","Opel","Ford"]}
```

### Get value by keys
```
GET /api/v1/values/{key}
```
 Parameter | Required | Description
-----------|----------|------------------------------------------------------
 key       | yes      | Cache key
 listIndex | no       | To get listIndex element for a list cache value
 dictKey   | no       | To get value by key from dict cache value
 
 Return code | Return value                                  | Description
-------------|-----------------------------------------------|---------------
 200         | {"value":"Mercury"}                           | Value
 400         | {"error":"Indicated value is not dictionary"} | if the dictKey param indicated for not dictionary value
 400         | {"error":"Indicated value is not list"}       | if the listIndex param indicated for not list value
 404         | {"error":"Cache item not found"}              | if the key does not exist
 
 _**Examples**_:
```
Request: curl -X GET http://localhost:8080/api/v1/values/planets
Response code: 200
Response body: {"value":{"planet1":"Mercury","planet2":"Venus","planet3":"Earth"}}

Request: curl -X GET http://localhost:8080/api/v1/values/cars?listIndex=1
Response code: 200
Response body: {"value":"Opel"}

Request: curl -X GET http://localhost:8080/api/v1/values/planets?dictKey=planet1
Response code: 200
Response body: {"value":"Mercury"}

Request: curl -X GET http://localhost:8080/api/v1/values/nonExistent
Response code: 404
Response body: {"error":"Cache item not found"}
```

### Get a list of available keys
```
GET /api/v1/keys
```

_**Examples**_:
```
Request: curl -X GET http://localhost:8080/api/v1/keys
Response code: 200
Response body: {"value":["planets","cars"]}
```

### Delete value by keys
```
DELETE /api/v1/values/{key}
```

 _**Examples**_:
```
Request: curl -X DELETE http://localhost:8080/api/v1/values/planets
Response code: 200
Response body: {"message":"Cache item deleted"}

Request: curl -X DELETE http://localhost:8080/api/v1/values/nonExistent
Response code: 404
Response body: {"error":"Cache item not found"}
```

### Set a timeout in seconds on key
```
PUT /api/v1/expire/{key}
```
 Return code | Return value                      | Description
-------------|-----------------------------------|------------------------------------------------------
 404         | {"error":"Cache item not found"}  | if key does not exist
 200         | {"message":"The timeout was set"} | if the timeout was set
 
 _**Examples**_:
```
Request: curl -X PUT -H 'content-type: application/json' -d 10 http://localhost:8080/api/v1/expire/nonExistent
Response code: 404
Response body: {"error":"Cache item not found"}

Request: curl -X PUT -H 'content-type: application/json' -d 10 http://localhost:8080/api/v1/expire/cars
Response code: 200
Response body: {"message":"The timeout was set"}
```

### Get the remaining time to live in seconds of a key that has a timeout
```
GET /api/v1/ttl/{key}
```

 Return code | Return value                      | Description
-------------|-----------------------------------|------------------
 200         | {"value":11}                      | TTL in seconds
 404         | {"error":"Cache item not found"}  | if the key does not exist or has no associated expire
   
 _**Examples**_:
```
Request: curl -X GET http://localhost:8080/api/v1/ttl/sixthMonth
Response code: 200
Response body: {"value":11}

Request: curl -X GET http://localhost:8080/api/v1/ttl/nonExistent
Response code: 404
Response body: {"error":"Cache item not found"}
```

### Persist cache data
```
POST /api/v1/persist
```

 Return code | Return value                      | Description
-------------|-----------------------------------|------------------
 200         | {"message":"Cache Data persisted"}| Cache Data persisted
 500         | {"error":"Error message"}         | Internal server error
 
 _**Examples**_:
```
Request: curl -X POST http://localhost:8080/api/v1/persist
Response code: 200
Response body: {"message":"Cache Data persisted"}
```

### Reload persisted cache data
```
POST /api/v1/reload
```

 Return code | Return value                      | Description
-------------|-----------------------------------|------------------
 200         | {"message":"Cache Data reloaded"} | Cache Data reloaded
 500         | {"error":"Error message"}         | Internal server error
 
 _**Examples**_:
```
Request: curl -X POST http://localhost:8080/api/v1/reload
Response code: 200
Response body: {"message":"Cache Data reloaded"}
```
