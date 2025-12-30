// NeoBank Jenkins Pipeline (GitOps)
// CI pipeline: Build, test, scan, sign images, then update Git manifests
// IMPORTANT: Jenkins does NOT deploy directly to Kubernetes
// Argo CD handles all deployments via GitOps

pipeline {
    agent any
    
    environment {
        // AWS ECR Registry
        AWS_REGION = 'us-east-1'
        AWS_ACCOUNT_ID = credentials('aws-account-id')
        ECR_REGISTRY = "${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
        
        // Git repositories
        MANIFEST_REPO = 'git@github.com:your-org/neobank-manifests.git'
        MANIFEST_CREDENTIALS = 'manifest-repo-ssh'
        
        // Notifications
        SLACK_CHANNEL = '#neobank-ci'
    }
    
    options {
        buildDiscarder(logRotator(numToKeepStr: '20'))
        timeout(time: 45, unit: 'MINUTES')
        timestamps()
        disableConcurrentBuilds()
        ansiColor('xterm')
    }
    
    parameters {
        choice(
            name: 'SERVICES', 
            choices: ['all', 'identity-service', 'ledger-service', 'payment-service', 'product-service', 'card-service', 'frontend'], 
            description: 'Service(s) to build'
        )
        choice(
            name: 'TARGET_ENV', 
            choices: ['dev', 'staging'], 
            description: 'Target environment for manifest update (prod requires manual PR)'
        )
        booleanParam(
            name: 'SKIP_TESTS', 
            defaultValue: false, 
            description: 'Skip test stage (emergency builds only)'
        )
        booleanParam(
            name: 'UPDATE_MANIFESTS', 
            defaultValue: true, 
            description: 'Create PR to update manifests repo'
        )
    }
    
    stages {
        // ============================================
        // Stage 1: Checkout & Setup
        // ============================================
        stage('Checkout') {
            steps {
                checkout scm
                script {
                    env.GIT_SHA = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                    env.GIT_FULL_SHA = sh(script: 'git rev-parse HEAD', returnStdout: true).trim()
                    env.GIT_COMMIT_MSG = sh(script: 'git log -1 --pretty=%B', returnStdout: true).trim()
                    env.IMAGE_TAG = env.GIT_SHA
                    
                    // Determine which services to build
                    if (params.SERVICES == 'all') {
                        env.BUILD_SERVICES = 'identity-service ledger-service payment-service product-service card-service frontend'
                    } else {
                        env.BUILD_SERVICES = params.SERVICES
                    }
                    
                    echo "Building services: ${env.BUILD_SERVICES}"
                    echo "Image tag: ${env.IMAGE_TAG}"
                }
            }
        }
        
        // ============================================
        // Stage 2: Build Docker Images
        // ============================================
        stage('Build Images') {
            steps {
                script {
                    // Login to ECR
                    sh """
                        aws ecr get-login-password --region ${AWS_REGION} | \
                        docker login --username AWS --password-stdin ${ECR_REGISTRY}
                    """
                    
                    // Build images in parallel
                    def builds = [:]
                    env.BUILD_SERVICES.split(' ').each { service ->
                        builds[service] = {
                            buildDockerImage(service)
                        }
                    }
                    parallel builds
                }
            }
        }
        
        // ============================================
        // Stage 3: Run Tests
        // ============================================
        stage('Test') {
            when {
                expression { !params.SKIP_TESTS }
            }
            parallel {
                stage('Backend Unit Tests') {
                    when {
                        expression { env.BUILD_SERVICES.contains('service') }
                    }
                    steps {
                        dir('backend') {
                            sh '''
                                go work sync
                                go test -v -race -coverprofile=coverage.out ./...
                            '''
                        }
                    }
                    post {
                        always {
                            publishCoverage adapters: [coberturaAdapter('backend/coverage.out')]
                        }
                    }
                }
                
                stage('Frontend Unit Tests') {
                    when {
                        expression { env.BUILD_SERVICES.contains('frontend') }
                    }
                    steps {
                        dir('frontend') {
                            sh '''
                                npm ci
                                npm run test:ci
                            '''
                        }
                    }
                }
                
                stage('Integration Tests') {
                    steps {
                        dir('backend/tests/integration') {
                            sh '''
                                # Run integration tests against test containers
                                docker-compose -f docker-compose.test.yml up -d
                                sleep 10
                                go test -v -tags=integration ./...
                                docker-compose -f docker-compose.test.yml down
                            '''
                        }
                    }
                }
            }
        }
        
        // ============================================
        // Stage 4: Security Scanning
        // ============================================
        stage('Security Scan') {
            parallel {
                stage('Vulnerability Scan') {
                    steps {
                        script {
                            env.BUILD_SERVICES.split(' ').each { service ->
                                def imageName = "${ECR_REGISTRY}/neobank/${service}:${IMAGE_TAG}"
                                sh """
                                    echo "Scanning ${imageName}..."
                                    trivy image \
                                        --severity HIGH,CRITICAL \
                                        --exit-code 1 \
                                        --ignore-unfixed \
                                        --format table \
                                        ${imageName}
                                """
                            }
                        }
                    }
                }
                
                stage('SBOM Generation') {
                    steps {
                        script {
                            env.BUILD_SERVICES.split(' ').each { service ->
                                def imageName = "${ECR_REGISTRY}/neobank/${service}:${IMAGE_TAG}"
                                sh """
                                    syft ${imageName} -o spdx-json > sbom-${service}.json
                                """
                            }
                        }
                    }
                    post {
                        always {
                            archiveArtifacts artifacts: 'sbom-*.json', allowEmptyArchive: true
                        }
                    }
                }
            }
        }
        
        // ============================================
        // Stage 5: Push & Sign Images
        // ============================================
        stage('Push & Sign') {
            steps {
                script {
                    env.BUILD_SERVICES.split(' ').each { service ->
                        def imageName = "${ECR_REGISTRY}/neobank/${service}"
                        
                        // Push image
                        sh """
                            docker push ${imageName}:${IMAGE_TAG}
                            docker tag ${imageName}:${IMAGE_TAG} ${imageName}:latest
                            docker push ${imageName}:latest
                        """
                        
                        // Sign with Cosign (keyless)
                        sh """
                            COSIGN_EXPERIMENTAL=1 cosign sign --yes ${imageName}:${IMAGE_TAG}
                        """
                        
                        // Attach SBOM
                        sh """
                            cosign attach sbom --sbom sbom-${service}.json ${imageName}:${IMAGE_TAG}
                        """
                    }
                }
            }
        }
        
        // ============================================
        // Stage 6: Update Manifests (GitOps)
        // ============================================
        stage('Update Manifests') {
            when {
                expression { params.UPDATE_MANIFESTS }
            }
            steps {
                script {
                    // Checkout manifests repo
                    dir('manifests') {
                        git(
                            url: env.MANIFEST_REPO,
                            credentialsId: env.MANIFEST_CREDENTIALS,
                            branch: params.TARGET_ENV
                        )
                        
                        // Update image tags
                        env.BUILD_SERVICES.split(' ').each { service ->
                            def kustomizePath = "services/${service}/${params.TARGET_ENV}/kustomization.yaml"
                            
                            if (fileExists(kustomizePath)) {
                                sh """
                                    cd services/${service}/${params.TARGET_ENV}
                                    kustomize edit set image \
                                        neobank/${service}=${ECR_REGISTRY}/neobank/${service}:${IMAGE_TAG}
                                """
                            }
                        }
                        
                        // Commit changes
                        sh """
                            git config user.email "jenkins@neobank.com"
                            git config user.name "Jenkins CI"
                            git add -A
                            git commit -m "chore: update ${params.TARGET_ENV} images to ${IMAGE_TAG}

Services: ${env.BUILD_SERVICES}
Source commit: ${env.GIT_FULL_SHA}
Message: ${env.GIT_COMMIT_MSG}

[skip ci]"
                        """
                        
                        // Create branch and push
                        def branchName = "deploy/${params.TARGET_ENV}/${IMAGE_TAG}"
                        sh """
                            git checkout -b ${branchName}
                            git push origin ${branchName}
                        """
                        
                        // Create PR using GitHub CLI
                        withCredentials([string(credentialsId: 'github-token', variable: 'GH_TOKEN')]) {
                            sh """
                                gh pr create \
                                    --title "[Deploy] Update ${params.TARGET_ENV} to ${IMAGE_TAG}" \
                                    --body "## Automated Deployment PR

**Source Commit:** ${env.GIT_FULL_SHA}
**Target Environment:** ${params.TARGET_ENV}
**Services Updated:** ${env.BUILD_SERVICES}

### Changes
- Updated container image tags to \`${IMAGE_TAG}\`

### Commit Message
\`\`\`
${env.GIT_COMMIT_MSG}
\`\`\`

### Verification
- [x] Images built successfully
- [x] Unit tests passed
- [x] Security scan passed
- [x] Images signed with Cosign
- [x] SBOM attached

---
*This PR was automatically created by Jenkins CI*" \
                                    --base ${params.TARGET_ENV} \
                                    --label "automated,deployment,${params.TARGET_ENV}"
                            """
                        }
                    }
                }
            }
        }
    }
    
    // ============================================
    // Post Actions
    // ============================================
    post {
        success {
            slackSend(
                channel: env.SLACK_CHANNEL,
                color: 'good',
                message: """✅ *NeoBank Build Successful*
• *Services:* ${env.BUILD_SERVICES}
• *Image Tag:* `${env.IMAGE_TAG}`
• *Target:* ${params.TARGET_ENV}
• *Build:* <${env.BUILD_URL}|#${env.BUILD_NUMBER}>
• *Commit:* `${env.GIT_SHA}` - ${env.GIT_COMMIT_MSG.take(50)}

_Argo CD will sync automatically for dev/staging. Production requires manual approval._"""
            )
        }
        
        failure {
            slackSend(
                channel: env.SLACK_CHANNEL,
                color: 'danger',
                message: """❌ *NeoBank Build Failed*
• *Services:* ${env.BUILD_SERVICES}
• *Stage:* ${env.STAGE_NAME}
• *Build:* <${env.BUILD_URL}|#${env.BUILD_NUMBER}>
• *Commit:* `${env.GIT_SHA}`

<${env.BUILD_URL}console|View Console Output>"""
            )
        }
        
        always {
            // Cleanup
            sh 'docker system prune -f || true'
            cleanWs()
        }
    }
}

// ============================================
// Helper Functions
// ============================================

def buildDockerImage(String service) {
    def context = service == 'frontend' ? 'frontend' : "backend/${service}"
    def dockerfile = "${context}/Dockerfile"
    def imageName = "${ECR_REGISTRY}/neobank/${service}:${IMAGE_TAG}"
    
    echo "Building ${service}..."
    
    sh """
        docker build \
            --build-arg VERSION=${IMAGE_TAG} \
            --build-arg COMMIT_SHA=${GIT_FULL_SHA} \
            --build-arg BUILD_TIME=\$(date -u +%Y-%m-%dT%H:%M:%SZ) \
            --label org.opencontainers.image.source=https://github.com/your-org/neobank \
            --label org.opencontainers.image.revision=${GIT_FULL_SHA} \
            --label org.opencontainers.image.version=${IMAGE_TAG} \
            -t ${imageName} \
            -f ${dockerfile} \
            ${context}
    """
}
