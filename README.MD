# USAGE

1. git clone

2. go build main.go

3. ./main


# Hypercloud-api-server API Specification
- Hypercloud-api-server에서 **제공하는 API 목록**
  - Prefix : http://{HYPERCLOUD5-API-SERVER-SERVICE}:{PORT}

## ALERT


#### Get
  - **Description** 
    생성 된 alert  Get 하는 서비스.
    필수 파라미터 = namespace
      
  - **RequestURL** 
  
    GET /alert
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      name : {NAME}  
      
      labelSelector  : {LABEL}  
      
      namespace : {NAMESPACE}  

  - **PathParam**

      x

  - **RequestBody**
  
      x
      
  - **ResponseBody**
  
      ex)
      ```json
      [
        {
          "kind": "AlertList",
          "apiVersion": "tmax.io/v1",   
          "metadata":{
            "selfLink": "/apis/tmax.io/v1/namespaces/default/alerts",
            "resourceVersion": "249653521",
            "creationTimestamp": null
           },
          "spec":{
            "kind": "warning",
            "name": "pod-1234",
            "resource": "podCPU",
            "message": "pod cpu usage 90%"
           }
        }
      ]
      ```
      
#### POST
  - **Description** 
  alert 리소스를 생성하는 service.
  alertmanager가 직접 호출 한다.
      
  - **RequestURL** 
  
    POST /alert
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      x

  - **PathParam**

      x

  - **RequestBody**
  
      ex)
      ```json
      [
        {
          "name" : "pod_name",  
          "alert" : "CPU_noti",
          "namespace" : "default",
          "message" : "cpu 20%"
         }
      ]
      ```
      
  - **ResponseBody**
  
      x


## METERING
schema,secret은 mysql-settings.yaml 에서 관리

#### Get
  - **Description** 
  
     현재 리소스 사용량을 조회하는 API
      
  - **RequestURL** 
  
    GET /metering
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      offset : {OFFSET}  
      
      limit : {LIMIT}  
      
      namespace : {NAMESPACE}  
      
      timeUnit : {TIMEUNIT} (hour, day, month, year) (default=day)  
      
      startTime : {STARTTIME}  
      
      endTime : {ENDTIME}  
      
      sort : {SORTS}  

  - **PathParam**

      x

  - **RequestBody**
  
      x
      
  - **ResponseBody**
  
      ex)
      ```json
      [
        {
          "id": "56e1c905-530c-4ece-8a00-a9bed2477607",
          "namespace": "hypercloud4-system",
          "cpu": 0,
          "memory": 1158947976,
          "storage": 10737418240,
          "gpu": 0,
          "publicIp": 2,
          "privateIp": 0,
          "trafficIn": 0,
          "trafficOut": 0,
          "meteringTime": "2021-01-15T00:00:00Z"
        },
        ...
      ]
      ```

## NAMESPACE

#### Get
  - **Description** 
  
    접근 가능한 namespace list를 조회하는 API
      
  - **RequestURL** 
  
    GET /namespace
    
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
      
#### PUT
  - **Description** 
  
  trial namespace 기간 연장 API
     
  - **RequestURL** 
  
    PUT /namespace
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      nsName : {NAMESPACE}  
      
      period : {ADDPERIOD}
      
  - **PathParam**

      x

  - **RequestBody**
  
      x
      
  - **ResponseBody**
  
      x
      
      
## NAMESPACECLAIM

#### Get
  - **Description** 
  
    접근 가능한 namespaceClaim list를 조회하는 API
      
  - **RequestURL** 
  
    GET /namespaceClaim
    
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
      
#### PUT
  - **Description** 
  
     Namespace 이름 중복 API
     
  - **RequestURL** 
  
    PUT /namespaceClaim
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      namespace : {NAMESPACE_NAME}  
      
  - **PathParam**

      x

  - **RequestBody**
  
      x
      
  - **ResponseBody**
  
      x
      
      
## USER


#### POST
  - **Description** 

      유저 회원가입시 유저에게 필요한 최초 클러스터 롤바인딩을 생성하는 API
      
  - **RequestURL** 
  
      POST /user
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      userId: {USER_ID}  

  - **PathParam**

      x

  - **RequestBody**
  
      x
    
  - **ResponseBody**
  
      x

#### DELETE
  - **Description** 

      유저 탈퇴시, 유저의 클러스터 롤바인딩을 제거하는 API
      
  - **RequestURL** 
  
      DELETE /user
    
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
  
    GET /version
    
  - **RequestHeader**
      
      x
    
  - **QueryParam**

      x

  - **PathParam**

      x

  - **RequestBody**
  
      x
      
  - **ResponseBody**
     
     eg)     
     ```json
      [
        {
        "name": "HyperCloud-Operator",
        "status": "Normal",
        "version": "b4.1.3"
        },
        {
        "name": "HyperCloud-Console",
        "status": "Normal",
        "version": "4.1.4"
        },
        ...
      ]
     ```
