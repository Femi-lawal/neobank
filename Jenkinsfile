// NeoBank Jenkins Pipeline
// Multi-stage CI/CD pipeline for building, testing, and deploying NeoBank

pipeline {
    agent any
    
    environment {
        DOCKER_REGISTRY = 'registry.neobank.com'
        DOCKER_CREDENTIALS = 'docker-registry-creds'
        KUBECONFIG = credentials('kubeconfig')
        SLACK_CHANNEL = '#deployments'
    }
    
    options {
        buildDiscarder(logRotator(numToKeepStr: '10'))
        timeout(time: 30, unit: 'MINUTES')
        timestamps()
        disableConcurrentBuilds()
    }
    
    parameters {
        choice(name: 'ENVIRONMENT', choices: ['dev', 'staging', 'prod'], description: 'Deployment environment')
        booleanParam(name: 'RUN_TESTS', defaultValue: true, description: 'Run tests')
        booleanParam(name: 'DEPLOY', defaultValue: true, description: 'Deploy after build')
        string(name: 'VERSION', defaultValue: '', description: 'Version tag (leave empty for git SHA)')
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
                script {
                    env.GIT_SHA = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                    env.VERSION = params.VERSION ?: env.GIT_SHA
                }
            }
        }
        
        stage('Build Backend Services') {
            parallel {
                stage('Identity Service') {
                    steps {
                        buildService('identity-service')
                    }
                }
                stage('Ledger Service') {
                    steps {
                        buildService('ledger-service')
                    }
                }
                stage('Payment Service') {
                    steps {
                        buildService('payment-service')
                    }
                }
                stage('Product Service') {
                    steps {
                        buildService('product-service')
                    }
                }
                stage('Card Service') {
                    steps {
                        buildService('card-service')
                    }
                }
            }
        }
        
        stage('Build Frontend') {
            steps {
                dir('frontend') {
                    sh '''
                        docker build -t ${DOCKER_REGISTRY}/frontend:${VERSION} .
                        docker tag ${DOCKER_REGISTRY}/frontend:${VERSION} ${DOCKER_REGISTRY}/frontend:latest
                    '''
                }
            }
        }
        
        stage('Run Tests') {
            when {
                expression { params.RUN_TESTS }
            }
            parallel {
                stage('Backend Tests') {
                    steps {
                        dir('backend') {
                            sh '''
                                cd shared-lib && go test ./... -v -coverprofile=coverage.out
                                cd ../identity-service && go test ./... -v
                                cd ../ledger-service && go test ./... -v
                                cd ../payment-service && go test ./... -v
                            '''
                        }
                    }
                    post {
                        always {
                            publishCoverage adapters: [coberturaAdapter('backend/**/coverage.out')]
                        }
                    }
                }
                stage('Frontend Tests') {
                    steps {
                        dir('frontend') {
                            sh '''
                                npm ci
                                npm run lint
                                npm run test:ci || true
                            '''
                        }
                    }
                }
                stage('Security Scan') {
                    steps {
                        sh '''
                            # Run security scan with Trivy
                            for service in identity-service ledger-service payment-service product-service card-service frontend; do
                                trivy image --severity HIGH,CRITICAL ${DOCKER_REGISTRY}/${service}:${VERSION} || true
                            done
                        '''
                    }
                }
            }
        }
        
        stage('Push Images') {
            steps {
                withCredentials([usernamePassword(credentialsId: env.DOCKER_CREDENTIALS, usernameVariable: 'DOCKER_USER', passwordVariable: 'DOCKER_PASS')]) {
                    sh '''
                        echo $DOCKER_PASS | docker login ${DOCKER_REGISTRY} -u $DOCKER_USER --password-stdin
                        
                        for service in identity-service ledger-service payment-service product-service card-service frontend; do
                            docker push ${DOCKER_REGISTRY}/${service}:${VERSION}
                            docker push ${DOCKER_REGISTRY}/${service}:latest
                        done
                    '''
                }
            }
        }
        
        stage('Deploy') {
            when {
                expression { params.DEPLOY }
            }
            steps {
                script {
                    def envFolder = params.ENVIRONMENT == 'prod' ? 'prod' : 'dev'
                    
                    sh """
                        export KUBECONFIG=${KUBECONFIG}
                        
                        # Update image tags in kustomization
                        cd k8s/overlays/${envFolder}
                        kustomize edit set image \\
                            neobank/identity-service=${DOCKER_REGISTRY}/identity-service:${VERSION} \\
                            neobank/ledger-service=${DOCKER_REGISTRY}/ledger-service:${VERSION} \\
                            neobank/payment-service=${DOCKER_REGISTRY}/payment-service:${VERSION} \\
                            neobank/product-service=${DOCKER_REGISTRY}/product-service:${VERSION} \\
                            neobank/card-service=${DOCKER_REGISTRY}/card-service:${VERSION} \\
                            neobank/frontend=${DOCKER_REGISTRY}/frontend:${VERSION}
                        
                        # Apply to Kubernetes
                        kubectl apply -k .
                        
                        # Wait for rollout
                        for deployment in identity-service ledger-service payment-service product-service card-service frontend; do
                            kubectl rollout status deployment/\${deployment} -n neobank --timeout=300s
                        done
                    """
                }
            }
        }
        
        stage('Health Check') {
            when {
                expression { params.DEPLOY }
            }
            steps {
                script {
                    def endpoints = [
                        'identity-service:8081',
                        'ledger-service:8082',
                        'payment-service:8083',
                        'product-service:8084',
                        'card-service:8085'
                    ]
                    
                    endpoints.each { endpoint ->
                        sh """
                            kubectl exec -n neobank deploy/\${endpoint.split(':')[0]} -- \\
                                wget --spider -q http://localhost:\${endpoint.split(':')[1]}/health
                        """
                    }
                }
            }
        }
    }
    
    post {
        success {
            slackSend(channel: env.SLACK_CHANNEL, color: 'good', 
                message: "✅ NeoBank ${params.ENVIRONMENT} deployment successful! Version: ${env.VERSION}")
        }
        failure {
            slackSend(channel: env.SLACK_CHANNEL, color: 'danger',
                message: "❌ NeoBank ${params.ENVIRONMENT} deployment failed! Build: ${env.BUILD_URL}")
        }
        always {
            cleanWs()
        }
    }
}

// Helper function to build a backend service
def buildService(String serviceName) {
    dir('backend') {
        sh """
            docker build -t ${DOCKER_REGISTRY}/${serviceName}:${VERSION} \\
                -f ${serviceName}/Dockerfile .
            docker tag ${DOCKER_REGISTRY}/${serviceName}:${VERSION} \\
                ${DOCKER_REGISTRY}/${serviceName}:latest
        """
    }
}
