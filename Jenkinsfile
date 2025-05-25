pipeline {
    agent any
    environment {
        IMAGE_NAME = "siwuai:${env.BUILD_ID}"
    }
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        stage('Build Docker Image') {
            steps {
                script {
                    // 构建 Docker 镜像
                    docker.build("${IMAGE_NAME}")
                }
            }
        }
        stage('Deploy to Server') {
            steps {
                sshagent(credentials: ['ssh-192-168-10-7']) {
                    sh '''
                    ssh user@192.168.10.7 << EOF
                    # 停止并删除旧容器（如果存在）
                    docker stop siwuai || true
                    docker rm siwuai || true
                    # 确保挂载目录存在
                    mkdir -p /home/siwu/ai/siwuaiservice/configs
                    mkdir -p /home/siwu/ai/siwuaiservice/logs
                    # 运行新容器，与手动部署一致
                    docker run -d --name siwuai -p 50051:50051 \
                        -v /home/siwu/ai/siwuaiservice/configs:/app/configs \
                        -v /home/siwu/ai/siwuaiservice/logs:/app/logs \
                        ${IMAGE_NAME}
                    # 清理旧镜像
                    docker image prune -f
                    EOF
                    '''
                }
            }
        }
    }
}