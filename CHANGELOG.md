# Hypercloud-api-server changelog!!
All notable changes to this project will be documented in this file.

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
-e 
<!-------------------- v5.0.11.8 start -------------------->
-e 
## Hypercloud-api-server 5.0.11.8 (2021. 05. 26. (수) 18:10:12 KST)
-e 
### Added
-e 
### Changed
-e 
### Fixed
-e 
### CRD yaml
-e 
### Etc
-e 
<!-------------------- v5.0.11.7 start -------------------->
-e 
## Hypercloud-api-server 5.0.11.7 (2021. 05. 26. (수) 17:50:17 KST)
-e 
### Added
-e 
### Changed
-e 
### Fixed
-e 
### CRD yaml
-e 
### Etc
-e 
<!-------------------- v5.0.11.6 start -------------------->
-e 
## Hypercloud-api-server 5.0.11.6 (2021. 05. 26. (수) 17:37:27 KST)
-e 
### Added
-e 
### Changed
  - [mod] ClusterManager 객채 스키마 변경에 따른 생성 로직 수정 (fakename 삭제) by chosangwon93
-e 
### Fixed
-e 
### CRD yaml
-e 
### Etc

<!-------------------- v5.0.11.5 start -------------------->

## Hypercloud-api-server 5.0.11.5 (Wed May 26 03:20:28 KST 2021)

### Added

### Changed
  - [mod] ClusterManager 객채 스키마 변경에 따른 생성 로직 수정 (fakename 삭제) by chosangwon93

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

<!-------------------- v5.0.11.4 start -------------------->

## Hypercloud-api-server 5.0.11.4 (Tue May 25 05:09:20 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc
-e 
<!-------------------- v5.0.11.2 start -------------------->
-e 
## Hypercloud-api-server 5.0.11.2 (2021. 05. 21. (금) 16:58:56 KST)
-e 
### Added
-e 
### Changed
-e 
### Fixed
-e 
### CRD yaml
-e 
### Etc
-e 
<!-------------------- v5.0.11.1 start -------------------->
-e 
## Hypercloud-api-server 5.0.11.1 (2021. 05. 21. (금) 16:11:23 KST)
-e 
### Added
-e 
### Changed
-e 
### Fixed
-e 
### CRD yaml
-e 
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

<!-------------------- v5.0.11.0 start -------------------->

## Hypercloud-api-server 5.0.11.0 (Thu May 20 08:19:45 KST 2021)

### Added

### Changed
  - [mod] version API가 잘못된 hyperauth 콜을 부를 경우 crash나던 현상 수정 by Seungwon Lee
  - [mod] hyperauth 정보 얻어오는 로직 추가 by Seungwon Lee

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
