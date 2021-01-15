# USAGE

1. git clone

2. go build main.go

3. ./main


# Hypercloud-api-server API Specification
- Hypercloud-api-server에서 **제공하는 API 목록**

## ALERT


#### Get
  - **Description** 
  
      
  - **RequestURL** 
  
    GET http://{HYPERCLOUD5-API-SERVER-SERVICE}:{PORT}/alert
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      name : {NAME}  
      
      label : {LABEL}  
      
      namespace : {NAMESPACE}  

  - **PathParam**

      x

  - **RequestBody**
  
      x
      
  - **ResponseBody**
  
      x
      
#### Create
  - **Description** 
  
      
  - **RequestURL** 
  
    POST http://{HYPERCLOUD5-API-SERVER-SERVICE}:{PORT}/alert
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      x

  - **PathParam**

      x

  - **RequestBody**
  
      
      
  - **ResponseBody**
  
      x


## METERING
schema,secret은 mysql-settings.yaml 에서 관리

#### Get
  - **Description** 
  
     namespace별 현재 리소스 사용량을 조회하는 API
      
  - **RequestURL** 
  
    GET http://{HYPERCLOUD5-API-SERVER-SERVICE}:{PORT}/metering
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      offset : {OFFSET}  
      
      limit : {LIMIT}  
      
      namespace : {NAMESPACE}  
      
      timeUnit : {TIMEUNIT} (hour, day, month, year) (default=day)  
      
      startTime : {STARTTIME}  
      
      endTime : {ENDTIME}  
      
      sorts : {SORTS}  

  - **PathParam**

      x

  - **RequestBody**
  
      x
      
  - **ResponseBody**
  
      x

## NAMESPACE

#### Get
  - **Description** 
  
    접근 가능한 namespace를 조회하는 API
      
  - **RequestURL** 
  
    GET http://{HYPERCLOUD5-API-SERVER-SERVICE}:{PORT}/namespace
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      userId : {USERID}  
      
      limit : {LIMIT}  
      
      labelSelector : {LABELSELECTOR}  
      
  - **PathParam**

      x

  - **RequestBody**
  
      x
      
  - **ResponseBody**
  
      x
      
#### AttributeUpdate
  - **Description** 
  
     
  - **RequestURL** 
  
    PUT http://{HYPERCLOUD5-API-SERVER-SERVICE}:{PORT}/namespace
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      nsName : {NAMESPACE}  
      
      addPeriod : {ADDPERIOD}
      
  - **PathParam**

      x

  - **RequestBody**
  
      x
      
  - **ResponseBody**
  
      x
      
      
## NAMESPACECLAIM

#### Get
  - **Description** 
  
     파라미터로 받은 유저의 namespaceClaim 권한을 조회하는 API
      
  - **RequestURL** 
  
    GET http://{HYPERCLOUD5-API-SERVER-SERVICE}:{PORT}/namespaceClaim
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      userId : {USERID}  
      
      limit : {LIMIT}  
      
      labelSelector : {LABELSELECTOR}  
      
  - **PathParam**

      x

  - **RequestBody**
  
      x
      
  - **ResponseBody**
  
      x
      
#### AttributeUpdate
  - **Description** 
  
     
  - **RequestURL** 
  
    PUT http://{HYPERCLOUD5-API-SERVER-SERVICE}:{PORT}/namespaceClaim
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      nsName : {NAMESPACE}  
      
  - **PathParam**

      x

  - **RequestBody**
  
      x
      
  - **ResponseBody**
  
      x
      
      
## USER


#### Create
  - **Description** 

      파라미터로 받은 유저의 클러스터 롤바인딩을 생성하는 API
      
  - **RequestURL** 
  
      POST http://{HYPERCLOUD5-API-SERVER-SERVICE}:{PORT}
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      userId: {USERID}  

  - **PathParam**

      x

  - **RequestBody**
  
      x
    
  - **ResponseBody**
  
      x

#### DELETE
  - **Description** 

      파라미터로 받은 유저의 클러스터 롤바인딩을 제거하는 API
      
  - **RequestURL** 
  
      DELETE http://{HYPERCLOUD5-API-SERVER-SERVICE}:{PORT}
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      userId : {USERID}  

  - **PathParam**

      x

  - **RequestBody**
  
      x
    
  - **ResponseBody**
  
      x


## VERSION
모듈의 version 정보에 관한 API  
verison 정보를 가져오기 위한 label정보는 hc-version configmap에 존재

#### Get
  - **Description** 
  
    마운트 된 configmap을 이용하여
    각 모듈의 version 정보를 얻어오는 API
      
  - **RequestURL** 
  
    GET http://{HYPERCLOUD5-API-SERVER-SERVICE}:{PORT}/version
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      x

  - **PathParam**

      x

  - **RequestBody**
  
      x
      
  - **ResponseBody**
  
      x
     
