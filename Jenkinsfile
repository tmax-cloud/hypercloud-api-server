node {
    def gitHubBaseAddress = "github.com"
	def hcBuildDir = "/var/lib/jenkins/workspace/hypercloud-api-server-golang"
	def imageBuildHome = "${hcBuildDir}"
	
	def scriptHome = "${hcBuildDir}/scripts"
	
	def gitHcAddress = "${gitHubBaseAddress}/tmax-cloud/hypercloud-api-server.git"

	def version = "${params.majorVersion}.${params.minorVersion}.${params.tinyVersion}.${params.hotfixVersion}"
	def preVersion = "${params.preVersion}"
	
	def imageTag = "b${version}"
				
	def userName = "taegeon_woo"
	def userEmail = "taegeon_woo@tmax.co.kr"
    
    stage('git pull & Go build') {
        dir(imageBuildHome){
            git branch: "${params.buildBranch}",
            credentialsId: '${userName}',
            url: "http://${gitHubBaseAddress}"

            // git pull
            sh "git checkout ${params.buildBranch}"
            sh "git config --global user.name ${userName}"
            sh "git config --global user.email ${userEmail}"
            sh "git config --global credential.helper store"
        
            sh "git fetch --all"
            sh "git reset --hard origin/${params.buildBranch}"
            sh "git pull origin ${params.buildBranch}"

            //go build
            sh "go build main.go"
        }
    }
    
	stage('image build & push'){
		dir(imageBuildHome){
		    sh "sudo docker build --tag tmaxcloudck/hypercloud-api-server:${imageTag} ."
			sh "sudo docker push tmaxcloudck/hypercloud-api-server:${imageTag}"
		}	
	}
	
	stage('make change log'){
        sh "sudo sh ${scriptHome}/hypercloud-changelog.sh ${version} ${preVersion}"
	}
	

	stage('git push'){
		dir ("${imageBuildHome}") {
			sh "git checkout ${params.buildBranch}"

			sh "git config --global user.name ${userName}"
			sh "git config --global user.email ${userEmail}"
			sh "git config --global credential.helper store"
			sh "git add -A"

			sh (script:'git commit -m "[Distribution] Hypercloud-api-server- ${version} " || true')
			sh "git tag v${version}"

			sh "sudo git push -u origin +${params.buildBranch}"
			sh "sudo git push origin v${version}"

			sh "git fetch --all"
			sh "git reset --hard origin/${params.buildBranch}"
			sh "git pull origin ${params.buildBranch}"
		}	
	}	   
}


