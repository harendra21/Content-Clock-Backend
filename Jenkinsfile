pipeline {
    agent any

    environment {
        IMAGE_NAME = 'content-clock-backend:latest'
        CONTAINER_NAME = 'content-clock-backend'
        DISCORD_WEBHOOK = 'https://discord.com/api/webhooks/1238396473419501599/E6amiHK6-y_kr-7pfNKXrEnI4edwO0TjgkiYcoe-TPPpvs6QJ3Bmnn5QGRvX9iclevLF'
    }

    stages {
        stage('Clone Repo') {
            steps {
                git url: 'https://github.com/harendra21/Content-Clock-Backend.git', branch: 'master'
            }
        }

        stage('Build Docker Image') {
            steps {
                script {
                    sh "docker build -t $IMAGE_NAME ."
                }
            }
        }

        stage('Stop Existing Container') {
            steps {
                script {
                    sh """
                        if [ \$(docker ps -q -f name=$CONTAINER_NAME) ]; then
                            docker stop $CONTAINER_NAME
                            docker rm $CONTAINER_NAME
                        fi
                    """
                }
            }
        }

        stage('Run New Container') {
            steps {
                script {
                    sh "docker run -d --name $CONTAINER_NAME --restart unless-stopped -p 8080:8080 -v $PWD/pbData:/pb/pb_data $IMAGE_NAME"
                }
            }
        }
    }

    post {
        success {
            sh """
                curl -H 'Content-Type: application/json' \
                -X POST \
                -d '{"content": "✅ Build SUCCESS: ${env.JOB_NAME} #${env.BUILD_NUMBER} - ${env.BUILD_URL}"}' \
                $DISCORD_WEBHOOK
            """
        }

        failure {
            sh """
                curl -H 'Content-Type: application/json' \
                -X POST \
                -d '{"content": "❌ Build FAILED: ${env.JOB_NAME} #${env.BUILD_NUMBER} - ${env.BUILD_URL}"}' \
                $DISCORD_WEBHOOK
            """
        }
    }
}