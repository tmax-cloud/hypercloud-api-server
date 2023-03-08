# Hypercloud-api-server changelog!!
All notable changes to this project will be documented in this file.

<!-------------------- v5.1.2.2 start -------------------->

## Hypercloud-api-server 5.1.2.2 (2023. 03. 08. (수) 10:45:51 KST)

### Added

### Changed
  - [mod] kubectl 이미지가 private registry를 지원하도록 환경변수에 PRIVATE_REGISTRY 추가 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.1.2.1 start -------------------->

## Hypercloud-api-server 5.1.2.1 (2023. 02. 28. (화) 13:00:56 KST)

### Added
  - [feat] clusterupdateclaim - cluster manager가 ready 일때만 승인되도록 예외 처리 추가 / log 개선 by sjoh0704
  - [feat] clusterupdateclaim - cluster manager type이 created일 때만 허용하도록 추가 by sjoh0704

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] clusterupdateclaim bug fix by sjoh0704

<!-------------------- v5.1.2.0 start -------------------->

## Hypercloud-api-server 5.1.2.0 (2023. 02. 23. (목) 15:07:54 KST)

### Added
  - [feat] add cluster update 승인/거절 api by sjoh0704
  - [feat] cluster update claim client 생성 및 list api 생성 by sjoh0704
  - [feat] dependency upgrade by sjoh0704
  - [feat] add cluster update claim router by sjoh0704

### Changed
  - [mod] Dockerfile go version update by sjoh0704

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.1.1.0 start -------------------->

## Hypercloud-api-server 5.1.1.0 (2023. 01. 18. (수) 19:34:45 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] kubectl GC 리팩토링 by Seungwon Lee

<!-------------------- v5.1.1.0 start -------------------->

## Hypercloud-api-server 5.1.1.0 (2023. 01. 18. (수) 18:39:35 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] kubectl GC 리팩토링 by Seungwon Lee

<!-------------------- v5.1.0.2 start -------------------->

## Hypercloud-api-server 5.1.0.2 (2023. 01. 03. (화) 15:09:42 KST)

### Added

### Changed
  - [mod] kubectl pod에 필요한 configmap을 보안을 위해 pod 생성 후에 삭제 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.1.0.1 start -------------------->

## Hypercloud-api-server 5.1.0.1 (2022. 12. 28. (수) 12:59:00 KST)

### Added

### Changed
  - [mod] kubectl 컨테이너가 default namespace를 바라보도록 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.1.0.0 start -------------------->

## Hypercloud-api-server 5.1.0.0 (2022. 12. 15. (목) 16:56:54 KST)

### Added

### Changed
  - [mod] kubectl GC cronjob을 leader election 로직 안으로 이동하여 하나의 pod만 수행하도록 수정 by Seungwon Lee
  - [mod] kubectl pod 생성 요청 시, 해당 pod에 대한 GET 권한 추가 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.38.2 start -------------------->

## Hypercloud-api-server 5.0.38.2 (2022. 12. 09. (금) 17:43:20 KST)

### Added

### Changed
  - [mod] kubectl pod가 Completed 상태인 경우에도 남은시간을 정상적으로 반환하도록 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.38.1 start -------------------->

## Hypercloud-api-server 5.0.38.1 (2022. 12. 08. (목) 14:23:05 KST)

### Added

### Changed
  - [mod] GET ~/kubectl이 남은 시간을 반환하도록 수정 by Seungwon Lee
  - [mod] kubectl 리소스 생성 시, 유저 이름에 포함된 _ 문자를 - 문자로 대체 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.38.0 start -------------------->

## Hypercloud-api-server 5.0.38.0 (2022. 12. 01. (목) 12:24:24 KST)

### Added
  - [feat] GET ~/kubectl 추가하여 이미지,timeout 정보 제공 by Seungwon Lee

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] kubectl API 관련 패키지 생성 by Seungwon Lee

<!-------------------- v5.0.37.0 start -------------------->

## Hypercloud-api-server 5.0.37.0 (2022. 11. 24. (목) 15:37:55 KST)

### Added
  - [feat] multi-operator package version upgrade 32=>36 by sjoh0704
  - [feat] cluster group 초대시 jwt-decode-auth용 service account 생성(임시 작업) by sjoh0704
  - [feat] cluster 초대 기능 사용시, jwt-decode-auth가 secret을 읽을 수 있도록 토큰 생성 방식 변경 by sjoh0704

### Changed
  - [mod] disable group invite by sjoh0704
  - [mod] clusterclaim, clusterregistration member Name cho => default by sjoh0704
  - [mod] cluster 초대시 oidc용 clusterrolebinding 복구 by sjoh0704
  - [mod] DELETE ~/kubectl 성공 시 response 보내도록 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc
  - [etc] version-config에 hyperauth url 비정상 기입시 로깅 by Seungwon Lee

<!-------------------- v5.0.36.0 start -------------------->

## Hypercloud-api-server 5.0.36.0 (2022. 11. 11. (금) 12:58:07 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.35.2 start -------------------->

## Hypercloud-api-server 5.0.35.2 (2022. 11. 11. (금) 11:07:38 KST)

### Added

### Changed

### Fixed
  - [ims][293625] clusterclaim rejected 상태일 떄 approved로 변경가능하도록 수정 by sjoh0704

### CRD yaml

### Etc

<!-------------------- v5.0.35.1 start -------------------->

## Hypercloud-api-server 5.0.35.1 (2022. 11. 09. (수) 18:54:48 KST)

### Added

### Changed
  - [mod] kubectl 이미지 bitnami/kubectl:1.19.16로 변경 by GitHub

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.35.0 start -------------------->

## Hypercloud-api-server 5.0.35.0 (2022. 11. 04. (금) 13:20:17 KST)

### Added
  - [feat] 콘솔에서 kubectl CLI 기능 제공을 위한 백엔드 기능 추가 by Seungwon Lee

### Changed
  - [mod] kubectl pod가 이미 있는 경우에도 200 OK 응답하도록 수정 by GitHub

### Fixed

### CRD yaml

### Etc
  - [etc] version.config를 최초 기동 시 한 번만 read 하도록 리팩토링 by Seungwon Lee

<!-------------------- v5.0.34.2 start -------------------->

## Hypercloud-api-server 5.0.34.2 (2022. 10. 14. (금) 09:24:35 KST)

### Added

### Changed
  - [mod] delete configmap from BindableResources by 2smin

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.34.1 start -------------------->

## Hypercloud-api-server 5.0.34.1 (2022. 09. 30. (금) 14:34:20 KST)

### Added

### Changed
  - [mod] add Secret to bindableResources by 2smin
  - [mod] bindableResource에 secret 추가 by 2smin
  - [mod] add redis kafka to GetBindableResources by 2smin

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.34.0 start -------------------->

## Hypercloud-api-server 5.0.34.0 (2022. 09. 21. (수) 16:05:25 KST)

### Added

### Changed
  - [mod] GET ~/bindableResources 에 redis, kafka 추가 변경 by 2smin

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.33.1 start -------------------->

## Hypercloud-api-server 5.0.33.1 (2022. 09. 08. (목) 18:20:48 KST)

### Added

### Changed
  - [mod] GET ~/event 버그 수정 및 kind 파라미터 복수개로 받을 수 있도록 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.33.0 start -------------------->

## Hypercloud-api-server 5.0.33.0 (2022. 09. 02. (금) 14:18:17 KST)

### Added

### Changed
  - [mod] leader election이 종료가 돼도  다시 실행되도록 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc
  - [etc] GetBindableResources 로직 수정 by 2smin

<!-------------------- v5.0.32.0 start -------------------->

## Hypercloud-api-server 5.0.32.0 (2022. 08. 24. (수) 12:47:54 KST)

### Added

### Changed
  - [mod] k8s event watch 기능 HA 고려하도록 수정 및 main.go init 과정 리팩토링 by Seungwon Lee
  - [mod] hypercloud multi-operator package 최신화 by SISILIA

### Fixed

### CRD yaml

### Etc
  - [etc] string 환경 변수 확인 방식을 strings.EqualFold 함수로 변경 by Seungwon Lee
  - [etc] 로그 레벨 지정 누락된 로그 수정 by Seungwon Lee
  - [etc] unused package 제거 by SISILIA

<!-------------------- v5.0.31.1 start -------------------->

## Hypercloud-api-server 5.0.31.1 (2022. 08. 16. (화) 17:12:58 KST)

### Added

### Changed
  - [mod] postgres에서 timescaledb로 DB 변경 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.31.0 start -------------------->

## Hypercloud-api-server 5.0.31.0 (2022. 08. 12. (금) 10:35:30 KST)

### Added
  - [feat] ServiceBinding - CustomResource List 불러오기 기능 추가 by 2smin
  - [feat] GET ~/event API 기능 추가 by Seungwon Lee
  - [feat] k8s event를 watch하여 DB에 저장하는 기능 추가 by Seungwon Lee

### Changed
  - [mod] GET ~/event 반환 형식을 events.v1 그룹에서 core.v1 그룹으로 변경 by Seungwon Lee
  - [mod] event UPDATE 시, DELETE 후 INSERT 하도록 로직 변경(hypertable 이슈) by Seungwon Lee

### Fixed

### CRD yaml

### Etc
  - [etc] bindableResources url 변경 by 2smin
  - [etc] func 이름 변경, http GET 만 받을 수 있도록 로직 변경 by 2smin
  - [etc] eventTime 예외처리 by Seungwon Lee
  - [etc] GET ~/metering 관련 함수 리팩토링 by Seungwon Lee

<!-------------------- v5.0.30.0 start -------------------->

## Hypercloud-api-server 5.0.30.0 (2022. 08. 02. (화) 17:08:22 KST)

### Added

### Changed
  - [mod] cluster manager 생성 로직 삭제 by seung
  - [mod] 버전 정규식 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.29.2 start -------------------->

## Hypercloud-api-server 5.0.29.2 (2022. 07. 22. (금) 18:39:25 KST)

### Added
  - [feat] 로그 레벨 기능 추가 by Seungwon Lee

### Changed
  - [mod] 로그레벨을 환경변수가 아닌 log-level 파라미터로 받도록 수정 by Seungwon Lee
  - [mod] start.sh이 아닌 main.go에서 LOG_LEVEL 파싱하도록 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.29.1 start -------------------->

## Hypercloud-api-server 5.0.29.1 (2022. 06. 30. (목) 17:48:24 KST)

### Added

### Changed
  - [mod] alert 관련 로직 제거 by Seungwon Lee

### Fixed

### CRD yaml

### Etc
  - [etc] K8sApiCaller.go의 함수가 error를 리턴하도록 리팩토링 by Seungwon Lee

<!-------------------- v5.0.29.0 start -------------------->

## Hypercloud-api-server 5.0.29.0 (2022. 06. 28. (화) 11:27:13 KST)

### Added

### Changed
  - [mod] grafana 관련 로직 제거 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.28.2 start -------------------->

## Hypercloud-api-server 5.0.28.2 (2022. 06. 22. (수) 19:27:53 KST)

### Added

### Changed
  - [mod] namespace spec에 " 문자가 들어간 경우의 에러 수정 및 NamespaceEvent 구조체 정의 by Seungwon Lee
  - [mod] namespace websocket 메시지 형식 변경 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.28.1 start -------------------->

## Hypercloud-api-server 5.0.28.1 (2022. 06. 20. (월) 13:42:47 KST)

### Added

### Changed
  - [mod] audit websocket 클라이언트 등록 시, 쿼리 파라미터도 받도록 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.28.0 start -------------------->

## Hypercloud-api-server 5.0.28.0 (2022. 06. 20. (월) 12:34:03 KST)

### Added
  - [feat] 웹소켓을 통해 namespace list를 받는 기능 추가, audit 웹소켓 기능 리팩토링 by Seungwon Lee

### Changed
  - [mod] websocket 클라이언트가 아무 메시지나 보낼 시, namespace list 반환 하도록 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.27.0 start -------------------->

## Hypercloud-api-server 5.0.27.0 (2022. 05. 24. (화) 15:44:47 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.26.19 start -------------------->

## Hypercloud-api-server 5.0.26.19 (2022. 05. 19. (목) 13:59:39 KST)

### Added
  - [feat] 초대된 cluster 멤버 목록 조회 api 추가 by sihyunglee823

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] main.go init() 리팩토링 by Seungwon Lee

<!-------------------- v5.0.26.18 start -------------------->

## Hypercloud-api-server 5.0.26.18 (2022. 05. 17. (화) 11:28:06 KST)

### Added
  - [feat] HA 지원을 위한 leader-election 적용 by Seungwon Lee

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] leader election 최대 유휴시간 10초로 감소 by Seungwon Lee
  - [etc] main.go 리팩토링 by Seungwon Lee

<!-------------------- v5.0.26.17 start -------------------->

## Hypercloud-api-server 5.0.26.17 (2022. 05. 08. (일) 17:50:27 KST)

### Added

### Changed
  - [mod] DB connection이 에러로 끊긴 이후, metering 데이터가 쌓이지 않던 버그 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.26.16 start -------------------->

## Hypercloud-api-server 5.0.26.16 (2022. 04. 15. (금) 16:13:36 KST)

### Added

### Changed
  - [mod] console 주소를 사용자가 입력한 subdomain을 이용하여 찾도록 수정 by SISILIA

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.26.15 start -------------------->

## Hypercloud-api-server 5.0.26.15 (2022. 04. 15. (금) 12:59:00 KST)

### Added

### Changed
  - [mod] metering 서비스 Metric 구조체의 Value 타입 변경 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.26.14 start -------------------->

## Hypercloud-api-server 5.0.26.14 (2022. 04. 13. (수) 16:22:03 KST)

### Added

### Changed
  - [mod] 클러스터 초대 메일 링크 수정 by SISILIA

### Fixed

### CRD yaml

### Etc
  - [etc] warning 코드 제거, log 오타 수정 by SISILIA
  - [etc] 오타 수정, warning code 제거 by SISILIA

<!-------------------- v5.0.26.13 start -------------------->

## Hypercloud-api-server 5.0.26.13 (2022. 04. 11. (월) 15:10:55 KST)

### Added
  - [feat] cluster 멤버, 그룹 조회 api 추가 by SISILIA
  - [feat] multi-operator 26버전과 호환을 위한 패키지 및 코드 추가 by SISILIA

### Changed
  - [mod] expire 코드 제거 by SISILIA
  - [mod] cluster invitation token 제거, invitation expire 로직 새로 추가 by SISILIA
  - [mod] clm list권한이 없는 경우에도 접근가능한 전체 clm을 조회할 수 있도록 변경 by SISILIA
  - [mod] console domain을 console service를 get하던 로직에서 env로 받게 변경 by SISILIA
  - [mod] code warning 수정 by SISILIA

### Fixed

### CRD yaml

### Etc
  - [etc] 오타 수정, warning 코드 수정 by SISILIA
  - [etc] code refactoring(워닝 코드 제거, deprecated 제거, 오타 수정) by SISILIA
  - [etc] 오타 수정 by SISILIA
  - [etc] 주석 추가 by SISILIA
  - [etc] code warning 수정 by SISILIA
  - [etc] fix log typo error by SISILIA
  - [etc] update gitignore by SISILIA

<!-------------------- v5.0.26.12 start -------------------->

## Hypercloud-api-server 5.0.26.12 (2022. 04. 07. (목) 18:30:30 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.26.11 start -------------------->

## Hypercloud-api-server 5.0.26.11 (2022. 02. 15. (화) 16:18:29 KST)

### Added

### Changed
  - [mod] metering type 변경 및 로그 출력 타입 수정 by Seungwon Lee
  - [mod] metering type 오류 수정 by Seungwon Lee

### Fixed
  - [ims][278080] GET ~/metering namespace 파라미터 수정 by Seungwon Lee

### CRD yaml

### Etc

<!-------------------- v5.0.26.10 start -------------------->

## Hypercloud-api-server 5.0.26.10 (2022. 01. 19. (수) 16:54:24 KST)

### Added

### Changed
  - [mod] GET /metering 수정, network I/O 쿼리 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.26.9 start -------------------->

## Hypercloud-api-server 5.0.26.9 (2022. 01. 03. (월) 10:50:16 KST)

### Added

### Changed
  - [mod] metering_year query문 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.26.8 start -------------------->

## Hypercloud-api-server 5.0.26.8 (2021. 12. 30. (목) 17:55:19 KST)

### Added

### Changed
  - [mod] metering INSERT 순서 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc
  - [etc] jwt package 공식 repo로 변경 by Seungwon Lee

<!-------------------- v5.0.26.7 start -------------------->

## Hypercloud-api-server 5.0.26.7 (2021. 12. 15. (수) 11:00:36 KST)

### Added

### Changed
  - [mod] audit batch 동작 오류로 보류 by Seungwon Lee
  - [mod] audit batch insert by Seungwon Lee

### Fixed

### CRD yaml

### Etc
  - [etc] clusterDataFactory DB connection refactoring by Seungwon Lee
  - [etc] DB connection refactoring by Seungwon Lee

<!-------------------- v5.0.26.6 start -------------------->

## Hypercloud-api-server 5.0.26.6 (2021. 12. 13. (월) 16:23:04 KST)

### Added

### Changed
  - [mod] sarama kafka 2.8.0 버전으로 업그레이드 by GitHub

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.26.5 start -------------------->

## Hypercloud-api-server 5.0.26.5 (2021. 12. 09. (목) 21:08:15 KST)

### Added

### Changed
  - [mod] kafka dns 오타 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.26.4 start -------------------->

## Hypercloud-api-server 5.0.26.4 (2021. 12. 09. (목) 17:51:54 KST)

### Added

### Changed
  - [mod] kafka DNS 1개로 축소 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.26.3 start -------------------->

## Hypercloud-api-server 5.0.26.3 (2021. 12. 08. (수) 16:44:52 KST)

### Added

### Changed
  - [mod] kafka DNS 주소 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.26.2 start -------------------->

## Hypercloud-api-server 5.0.26.2 (2021. 12. 08. (수) 11:28:58 KST)

### Added
  - [feat] GetAuditByJson 기능 추가 by Seungwon Lee
  - [feat] /cloudCredential api 추가 by Seungwon Lee

### Changed
  - [mod] kafka DNS 주소만 고려 by Seungwon Lee
  - [mod] insert audit json body 기능 비활성화 by Seungwon Lee
  - [mod] 다중 key 조회 가능하도록 수정 by Seungwon Lee
  - [mod] GET /audit/json 으로 변경 by Seungwon Lee
  - [mod] json_body insert 기능 추가 by Seungwon Lee
  - [mod] json body return 값 파싱 by Seungwon Lee
  - [mod] KAFKA_ENABLED 변수 추가 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.26.1 start -------------------->

## Hypercloud-api-server 5.0.26.1 (2021. 12. 06. (월) 15:26:16 KST)

### Added

### Changed
  - [mod] kafka DNS 주소만 고려 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.26.0 start -------------------->

## Hypercloud-api-server 5.0.26.0 (2021. 11. 04. (목) 13:57:24 KST)

### Added

### Changed
  - [mod] 인증서 방식 변경, kafka_enabled 분기 추가 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.25.2 start -------------------->

## Hypercloud-api-server 5.0.25.2 (2021. 11. 11. (목) 17:30:04 KST)

### Added

### Changed
  - [mod] KAFKA_ENABLED 변수 추가 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.25.1 start -------------------->

## Hypercloud-api-server 5.0.25.1 (2021. 10. 21. (목) 17:06:57 KST)

### Added

### Changed

### Fixed
  - [ims][271798] service instance parameter 공백 버그 수정 by soohwan kim

### CRD yaml

### Etc
  - [etc] undo build v5.0.25.6 by soohwan kim

<!-------------------- v5.0.25.0 start -------------------->

## Hypercloud-api-server 5.0.25.0 (2021. 08. 27. (금) 18:52:08 KST)

### Added
  - [feat] vsphere cluster claim을 위한 struct 추가 by soohwan kim

### Changed
  - [mod] postgres 연결 실패 시, pod 다운 되지 않도록 수정 by Seungwon Lee
  - [mod] 미사용 API 삭제 ~/awscost by Seungwon Lee

### Fixed

### CRD yaml

### Etc
  - [etc] go.mod multi-operator repo 버전 업데이트 by soohwan kim

<!-------------------- v5.0.24.0 start -------------------->

## Hypercloud-api-server 5.0.24.0 (2021. 08. 19. (목) 17:01:25 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.23.0 start -------------------->

## Hypercloud-api-server 5.0.23.0 (2021. 08. 12. (목) 13:03:29 KST)

### Added

### Changed
  - [mod] kafka 비정상 동작으로 인해 consumer group 등록 실패 시, 1분 후에 재시도 하도록 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.22.0 start -------------------->

## Hypercloud-api-server 5.0.22.0 (2021. 08. 05. (목) 15:14:49 KST)

### Added

### Changed
  - [mod] cluster 생성시 cluster manger creator annotation이 service account가 아닌 계정명으로 찍히도록 변경 by soohwan kim

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.21.0 start -------------------->

## Hypercloud-api-server 5.0.21.0 (2021. 07. 29. (목) 15:30:23 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.20.0 start -------------------->

## Hypercloud-api-server 5.0.20.0 (2021. 07. 22. (목) 14:17:03 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.19.1 start -------------------->

## Hypercloud-api-server 5.0.19.1 (2021. 07. 15. (목) 18:16:19 KST)

### Added

### Changed
  - [mod] nullString 에러 해결 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.19.0 start -------------------->

## Hypercloud-api-server 5.0.19.0 (2021. 07. 15. (목) 17:18:59 KST)

### Added

### Changed
  - [mod] cluster에 그룹 초대할 때 rolebinding 생성 에러 문제 해결 by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] refactor by chosangwon93

<!-------------------- v5.0.18.4 start -------------------->

## Hypercloud-api-server 5.0.18.4 (2021. 07. 13. (화) 12:51:06 KST)

### Added

### Changed

### Fixed
  - [ims][265974] hypercloud mutator 수정 후 nsc 승인 시 ns 생성 안되는 문제 해결 by chosangwon93

### CRD yaml

### Etc

<!-------------------- v5.0.18.3 start -------------------->

## Hypercloud-api-server 5.0.18.3 (2021. 07. 12. (월) 17:51:44 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.18.2 start -------------------->

## Hypercloud-api-server 5.0.18.2 (2021. 07. 12. (월) 14:11:34 KST)

### Added

### Changed
  - [mod] 클러스터 초대 수락 시 redirect url 변경 by chosangwon93
  - [mod] 클러스터에 초대된 사용자 삭제 시에 해당 ns에 사용중인 클러스터가 남아있는데도 불구하고 ns-get rolebinding 삭제되는 문제 해결 by chosangwon93

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.18.1 start -------------------->

## Hypercloud-api-server 5.0.18.1 (2021. 07. 09. (금) 17:06:45 KST)

### Added

### Changed
  - [mod] 카프라 groupid를 변수로 받도록 수정 by chosangwon93

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.18.0 start -------------------->

## Hypercloud-api-server 5.0.18.0 (2021. 07. 09. (금) 09:36:55 KST)

### Added
  - [feat] audit verb list api 추가 by chosangwon93

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.17.5 start -------------------->

## Hypercloud-api-server 5.0.17.5 (2021. 07. 08. (목) 13:29:58 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.17.3 start -------------------->

## Hypercloud-api-server 5.0.17.3 (2021. 07. 07. (수) 18:06:51 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.17.2 start -------------------->

## Hypercloud-api-server 5.0.17.2 (2021. 07. 07. (수) 17:31:17 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.17.1 start -------------------->

## Hypercloud-api-server 5.0.17.1 (2021. 07. 05. (월) 10:54:20 KST)

### Added

### Changed
  - [mod] 클러스터에 사용자 초대시 발생하는 에러 수정 by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] 리팩토링 by chosangwon93

<!-------------------- v5.0.17.0 start -------------------->

## Hypercloud-api-server 5.0.17.0 (2021. 07. 01. (목) 17:37:48 KST)

### Added
  - [feat] 클러스터 등록 기능 추가 by chosangwon93

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.16.1 start -------------------->

## Hypercloud-api-server 5.0.16.1 (2021. 07. 01. (목) 10:45:05 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.16.0 start -------------------->

## Hypercloud-api-server 5.0.16.0 (2021. 06. 25. (금) 11:06:55 KST)

### Added

### Changed
  - [mod] clusterclaim 수락을 통해 clustermanager 생성 시 label에 타입이 생성이라는 것을 나타내도록 수정 by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] dockerfile 수정을 통해 이미지 용량 축소 by Seungwon Lee
  - [etc]version API 로그 수정 by Seungwon Lee
  - [etc] versionHandler 리팩토링 및 결과 로그 출력 by Seungwon Lee

<!-------------------- v5.0.15.0 start -------------------->

## Hypercloud-api-server 5.0.15.0 (2021. 06. 17. (목) 15:09:12 KST)

### Added

### Changed
  - [mod] version parsing 정규식 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.14.0 start -------------------->

## Hypercloud-api-server 5.0.14.0 (2021. 06. 10. (목) 20:17:17 KST)

### Added
  - [feat] audit 리스트 서비스에서 ns-owner인 사용자만 해당 NS에 대한 감사기록을 볼 수 있도록 인가 로직 추가 by chosangwon93

### Changed
  - [mod] multi-operator가 클러스터 삭제 시 클러스터 정보를 db에서 지우기 위한 api call 추가 by chosangwon93
  - [mod] cluster에 사용자 초대하는 메일에 권한에 대한 부분 추가 by chosangwon93
  - [mod] 클러스터 클레임 승인 시 클러스터 이름 중복체크하는 로직 버그 수정 by chosangwon93
  - [mod] audit 리소스 목록을 apigroup/version과 상관없이 리소스 kind로만 반화도록 수정 by chosangwon93
  - [mod] clustermanager가 어떤 클레임으로부터 생성되었는지 표시하기 위해서 clustermanager 생성 시 lable에 claim 이름을 추가 by chosangwon93
  - [mod] audit resource 목록을 조회 할 때마다 중복으로 목록이 쌓이는 문제 해결 by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] status code change by chosangwon93
  - [etc] audit 불필요한 로직 제거 by chosangwon93

<!-------------------- v5.0.13.1 start -------------------->

## Hypercloud-api-server 5.0.13.1 (2021. 06. 03. (목) 18:53:50 KST)

### Added

### Changed
  - [mod] audit 리소스 리스트 서비스 및 리스트 시 인가 기능 추가 by chosangwon93

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.13.0 start -------------------->

## Hypercloud-api-server 5.0.13.0 (2021. 06. 03. (목) 17:20:53 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.12.0 start -------------------->

## Hypercloud-api-server 5.0.12.0 (2021. 05. 27. (목) 14:58:18 KST)

### Added
  - [feat] add resource list service for audit by chosangwon93

### Changed
  - [mod] 네임스페이스 내에서 클러스터 이름 중복을 허용하지 않도록 정책이 변경됨에 따라 로직 수정 by chosangwon93

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.11.6 start -------------------->

## Hypercloud-api-server 5.0.11.6 (2021. 05. 26. (수) 17:37:27 KST)
 
### Added
 
### Changed
  - [mod] ClusterManager 객채 스키마 변경에 따른 생성 로직 수정 (fakename 삭제) by chosangwon93
 
### Fixed
 
### CRD yaml
 
### Etc

<!-------------------- v5.0.11.5 start -------------------->

## Hypercloud-api-server 5.0.11.5 (Wed May 26 03:20:28 KST 2021)

### Added

### Changed
  - [mod] ClusterManager 객채 스키마 변경에 따른 생성 로직 수정 (fakename 삭제) by chosangwon93

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.11.4 start -------------------->

## Hypercloud-api-server 5.0.11.4 (Tue May 25 05:09:20 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.11.3 start -------------------->

## Hypercloud-api-server 5.0.11.3 (Wed May 26 02:25:43 KST 2021)

### Added

### Changed
  - [mod] ClusterManager 객채 스키마 변경에 따른 생성 로직 수정 (fakename 삭제) by chosangwon93

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.11.2 start -------------------->

## Hypercloud-api-server 5.0.11.2 (2021. 05. 21. (금) 16:58:56 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.11.1 start -------------------->

## Hypercloud-api-server 5.0.11.1 (2021. 05. 21. (금) 16:11:23 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.11.0 start -------------------->

## Hypercloud-api-server 5.0.11.0 (Thu May 20 08:19:45 KST 2021)

### Added

### Changed
  - [mod] version API가 잘못된 hyperauth 콜을 부를 경우 crash나던 현상 수정 by Seungwon Lee
  - [mod] hyperauth 정보 얻어오는 로직 추가 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.10.0 start -------------------->

## Hypercloud-api-server 5.0.10.0 (Thu May 13 08:39:42 KST 2021)

### Added

### Changed
  - [mod] hyperauth 유저 탈퇴시, CRB/RB 삭제 함수 추가 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.9.0 start -------------------->

## Hypercloud-api-server 5.0.9.0 (Thu May  6 09:35:09 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.8.0 start -------------------->

## Hypercloud-api-server 5.0.8.0 (Fri Apr 30 08:56:29 KST 2021)

### Added

### Changed
  - [mod] kafka 주소 사용자에게 입력 받는 로직 by Seungwon Lee
  - [mod] kafka 없어도 서버 다운 안되는 방어 로직 추가 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.7.0 start -------------------->

## Hypercloud-api-server 5.0.7.0 (Thu Apr 22 10:29:21 KST 2021)

### Added
  - [feat] create cluster in claim admit API by chosangwon93

### Changed
  - [mod] seperate cluster list api by chosangwon93
  - [mod] bugfix sql query by chosangwon93
  - [mod] api url by chosangwon93
  - [mod] api url by chosangwon93
  - [mod] change resource scope by chosangwon93

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.6.0 start -------------------->

## Hypercloud-api-server 5.0.6.0 (Thu Apr  8 09:39:24 KST 2021)

### Added

### Changed
  - [mod] seperate cluster list api by chosangwon93
  - [mod] bugfix sql query by chosangwon93
  - [mod] api url by chosangwon93
  - [mod] api url by chosangwon93
  - [mod] change resource scope by chosangwon93

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.5.1 start -------------------->

## Hypercloud-api-server 5.0.5.1 (Mon Apr  5 02:57:55 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.3.1 start -------------------->

## Hypercloud-api-server 5.0.3.1 (Mon Apr  5 02:44:46 KST 2021)

### Added

### Changed
  - [mod] change resource scope by chosangwon93

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.5.0 start -------------------->

## Hypercloud-api-server 5.0.5.0 (Thu Apr  1 08:53:25 KST 2021)

### Added
  - [feat] deleteCRB/RB 추가 by Seungwon Lee

### Changed
  - [mod] metering merge 버그 수정 by GitHub
  - [mod] awscostHandler 리팩토링 by Seungwon Lee
  - [mod] 세션 생성 방식 변경 by Seungwon Lee
  - [mod] 파라미터 형식 수정 by Seungwon Lee
  - [mod] DB merge 조건문 버그 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc
  - [etc] combine to delete /user by Seungwon Lee

<!-------------------- v5.0.4.0 start -------------------->

## Hypercloud-api-server 5.0.4.0 (Thu Mar 25 09:28:54 KST 2021)

### Added
  - [feat] awscost api by Seungwon Lee

### Changed
  - [mod] datafactory by chosangwon93
  - [mod] 정규식 수정 by Seungwon Lee
  - [mod] output 변경 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.3.0 start -------------------->

## Hypercloud-api-server 5.0.3.0 (Thu Mar 18 17:51:05 KST 2021)

### Added

### Changed
  - [mod] remove cluster member info from status by chosangwon93

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.2.2 start -------------------->

## Hypercloud-api-server 5.0.2.2 (Tue Mar 16 17:26:01 KST 2021)

### Added

### Changed
  - [mod] version directory 구조 변경 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.2.1 start -------------------->

## Hypercloud-api-server 5.0.2.1 (Tue Mar 16 12:37:06 KST 2021)

### Added
  - [feat] deleteClaim api 추가 by Seungwon Lee

### Changed
  - [mod] deleteClaim api 이름 수정 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.2.0 start -------------------->

## Hypercloud-api-server 5.0.2.0 (Thu Mar 11 19:31:21 KST 2021)

### Added

### Changed
  - [mod] manage cluster member using db by chosangwon93

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.1.2 start -------------------->

## Hypercloud-api-server 5.0.1.2 (Wed Mar 10 11:37:57 KST 2021)

### Added

### Changed
  - [mod] Insert문 수정 by Seungwon Lee
  - [mod] metering DB postgres로 변경 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.1.1 start -------------------->

## Hypercloud-api-server 5.0.1.1 (Tue Mar  9 11:34:52 KST 2021)

### Added

### Changed
  - [mod] 빌드시스템 수정, kafka 주석 해제 by Seungwon Lee
  - [mod] log 제목 변경, 젠킨스 삭제 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.1.0 start -------------------->

## Hypercloud-api-server 5.0.1.0 (Mon Mar  8 15:53:34 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.43 start -------------------->

## Hypercloud-api-server 5.0.0.43 (Mon Mar  8 15:39:04 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.42 start -------------------->

## Hypercloud-api-server 5.0.0.42 (Mon Mar  8 15:32:22 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.40 start -------------------->

## Hypercloud-api-server 5.0.0.40 (Mon Mar  8 14:52:29 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.25 start -------------------->

## Hypercloud-api-server 5.0.0.25 (Mon Mar  8 13:51:22 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.24 start -------------------->

## Hypercloud-api-server 5.0.0.24 (Tue Mar  2 17:37:27 KST 2021)

### Added
  - [feat] 유저 삭제시 claim 삭제 로직 완성 by Seungwon Lee
  - [feat] 유저 삭제시 claim 삭제  초안 by Seungwon Lee

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] refactor by chosangwon93

<!-------------------- v5.0.0.23 start -------------------->

## Hypercloud-api-server 5.0.0.23 (Wed Feb 17 14:29:33 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.22 start -------------------->

## Hypercloud-api-server 5.0.0.22 (Fri Feb  5 06:10:15 KST 2021)

### Added
  - [feat] add updateMemberRole for remote cluster by cho

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.21 start -------------------->

## Hypercloud-api-server 5.0.0.21 (Thu Feb  4 12:33:30 KST 2021)

### Added

### Changed
  - [mod] add owner annotations by cho

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.20 start -------------------->

## Hypercloud-api-server 5.0.0.20 (Thu Feb  4 11:09:19 KST 2021)

### Added
  - [feat] create claim by cho

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.19 start -------------------->

## Hypercloud-api-server 5.0.0.19 (Thu Feb  4 10:09:05 KST 2021)

### Added

### Changed
  - [mod] remove group by cho
  - [mod] log align by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.18 start -------------------->

## Hypercloud-api-server 5.0.0.18 (Thu Feb  4 05:21:11 KST 2021)

### Added

### Changed
  - [mod] remoteClusterSet bug fix by cho

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.17 start -------------------->

## Hypercloud-api-server 5.0.0.17 (Thu Feb  4 05:15:39 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.16 start -------------------->

## Hypercloud-api-server 5.0.0.16 (Thu Feb  4 05:04:21 KST 2021)

### Added

### Changed
  - [mod] metering log 출력방식 변경 by Seungwon Lee

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.15 start -------------------->

## Hypercloud-api-server 5.0.0.15 (Thu Feb  4 04:37:01 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.14 start -------------------->

## Hypercloud-api-server 5.0.0.14 (Thu Feb  4 02:37:05 KST 2021)

### Added

### Changed
  - [mod] add group by cho

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.13 start -------------------->

## Hypercloud-api-server 5.0.0.13 (Mon Feb  1 09:55:52 KST 2021)

### Added

### Changed
  - [mod] add group query by cho
  - [mod] limit parameter bug fix by Seungwon Lee

### Fixed

### CRD yaml

### Etc
  - [etc] merge hypercloud webhook by cho

<!-------------------- v5.0.0.12 start -------------------->

## Hypercloud-api-server 5.0.0.12 (Thu Jan 28 05:51:04 KST 2021)

### Added

### Changed
  - [mod] create logs dir by cho

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.11 start -------------------->

## Hypercloud-api-server 5.0.0.11 (Wed Jan 27 09:43:18 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.10 start -------------------->

## Hypercloud-api-server 5.0.0.10 (Wed Jan 27 09:25:49 KST 2021)

### Added
  - [feat] topicEvent struct 생성 by dnxorjs1

### Changed
  - [mod] topicEvent struct 오류 수정 by dnxorjs1

### Fixed

### CRD yaml

### Etc
  - [etc] delete kafaka consumer by cho
  - [etc] add default user role by cho

<!-------------------- v5.0.0.9 start -------------------->

## Hypercloud-api-server 5.0.0.9 (Tue Jan 26 05:50:50 KST 2021)

### Added

### Changed
  - [mod] 빌드시스템 수정 by dnxorjs1

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.8 start -------------------->

## Hypercloud-api-server 5.0.0.8 (Tue Jan 26 05:35:16 KST 2021)

### Added
  - [feat] kafka consumer 구현완료2 by dnxorjs1
  - [feat] kafka consumer 구현완료 by dnxorjs1

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.7 start -------------------->

## Hypercloud-api-server 5.0.0.7 (Tue Jan 26 02:56:04 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.6 start -------------------->

## Hypercloud-api-server 5.0.0.6 (Tue Jan 26 01:54:47 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.5 start -------------------->

## Hypercloud-api-server 5.0.0.5 (Mon Jan 25 14:26:43 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.4 start -------------------->

## Hypercloud-api-server 5.0.0.4 (Mon Jan 25 14:20:36 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.3 start -------------------->

## Hypercloud-api-server 5.0.0.3 (Mon Jan 25 14:14:11 KST 2021)

### Added

### Changed
  - [mod] return empty list by cho
  - [mod] change insertMeteringData() func by Seungwon Lee
  - [mod] change resource configuration by Seungwon Lee
  - [mod] change ResourceQuota by Seungwon Lee
  - [mod] integrate install_yaml by Seungwon Lee
  - [mod] change error log by Seungwon Lee
  - [mod] change role to clusterrole by cho
  - [mod] change DB URI by Seungwon Lee
  - [mod] change version configmap path and metering-prometheus connection by Seungwon Lee
  - [mod] change scope of clusterclaim and clustermanager by cho

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.2 start -------------------->

## Hypercloud-api-server 5.0.0.2 (Wed Jan 20 10:21:45 KST 2021)

### Added
  - [feat] jenkinsfile by taegeon_woo

### Changed

### Fixed

### CRD yaml

### Etc

<!-------------------- v5.0.0.1 start -------------------->

## Hypercloud-api-server 5.0.0.1 (Wed Jan 20 10:13:29 KST 2021)
asdf
### Added

### Changed

### Fixed

### CRD yaml

### Etc
