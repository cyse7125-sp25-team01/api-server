pipeline {
    agent any
    stages {
        stage('Check Commit Message') {
            steps {
                script {
                    def commitMessage = sh(script: "git log -1 --pretty=%B", returnStdout: true).trim()
                    echo "Commit Message: ${commitMessage}"
                    sh """
                    echo '${commitMessage}' | npx commitlint --extends '@commitlint/config-conventional'
                    """
                }
            }
        }       
    }
}