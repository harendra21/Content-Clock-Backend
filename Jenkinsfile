pipeline {
    agent any

    environment {
        IMAGE_NAME = 'content-clock-backend:latest'
        CONTAINER_NAME = 'content-clock-backend'
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
                    sh "docker run -d --name $CONTAINER_NAME -p 8080:8080 -v $PWD/pbData:/pb/pb_data $IMAGE_NAME"
                }
            }
        }
    }
}