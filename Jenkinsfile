pipeline {
    agent {
        dockerfile{
            dir 'build-agent'
            args '-v /var/run/docker.sock:/var/run/docker.sock'
        }
    }
    environment {
        TF_VAR_terraform_state_bucket_key     = credentials('TF_VAR_terraform_state_bucket_key')
        TF_VAR_terraform_state_bucket_secret     = credentials('TF_VAR_terraform_state_bucket_secret')
        TF_VAR_terraform_state_bucket_region     = credentials('TF_VAR_terraform_state_bucket_region')
        PUB_SSH_KEY_PATH     = credentials('484e713f-b32f-40b4-8c07-5768d351fb0e')
    }
    stages {
        stage('Build') {
            steps {
                sh 'docker version'
                sh 'deploy'
            }
        }
    }

    post {
        always {
            emailext (
                subject: "Jenkins job '${env.JOB_NAME}' #'${env.BUILD_NUMBER}': '${currentBuild.currentResult}'",
                body: """<p>Jenkins job: <strong>'${env.JOB_NAME}' #'${env.BUILD_NUMBER}'</strong> has completed with the result: <strong>'${currentBuild.currentResult}'</strong></p>
                <p>Please find more details at <a href="'${env.BUILD_URL}'">'${env.JOB_NAME}' #'${env.BUILD_NUMBER}'</a></p>""",
                mimeType: 'text/html',
                from: 'strengthgadget@gmail.com',
                to: 'i.need.notifications@gmail.com'
            )
        }
    }
}
