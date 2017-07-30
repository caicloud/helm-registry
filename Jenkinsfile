def IMAGE_TAG = "caicloud/helm-registry:${params.imageTag}"

podTemplate(
    cloud: 'dev-cluster',
    namespace: 'kube-system',
    name: 'helm-registry',
    label: 'helm-registry',
    idleMinutes: 1440,
    containers: [
        containerTemplate(
            name: 'jnlp',
            image: "cargo.caicloud.io/circle/jnlp:2.62",
            alwaysPullImage: true,
            command: '',
            args: '${computer.jnlpmac} ${computer.name}',
        ),
        containerTemplate(
            name: 'dind', 
            image: "cargo.caicloud.io/caicloud/docker:17.03-dind", 
            alwaysPullImage: true,
            ttyEnabled: true,
            command: '', 
            args: '--host=unix:///home/jenkins/docker.sock',
            privileged: true,
        ),
        containerTemplate(
            name: 'golang',
            image: "cargo.caicloud.io/caicloud/golang-docker:1.8.1-17.05",
            alwaysPullImage: true,
            ttyEnabled: true,
            command: '',
            args: '',
            envVars: [
                containerEnvVar(key: 'DOCKER_HOST', value: 'unix:///home/jenkins/docker.sock'),
                containerEnvVar(key: 'DOCKER_API_VERSION', value: '1.26'),
                containerEnvVar(key: 'WORKDIR', value: '/go/src/github.com/caicloud/helm-registry')
            ],
        ),
    ]
) {
    node('helm-registry') {
        stage('Checkout') {
            checkout scm
        }
        container('golang') {
            ansiColor('xterm') {

                stage("Complie") {
                    sh('''
                        set -e 
                        mkdir -p $(dirname ${WORKDIR})
                        rm -rf ${WORKDIR}
                        ln -sf $(pwd) ${WORKDIR}
                        cd ${WORKDIR}
                        make registry
                    ''')
                }

                stage('Run e2e test') {
                    if (!params.integration) {
                        echo "skip integration"
                        return
                    }
                    sh('''
                        set -e
                        cd ${WORKDIR}
                        make test
                    ''')
                }
            }

            stage("Build image and publish") {
                if (!params.publish) {
                    echo "skip publish"
                    return
                }
                sh "docker build -t ${IMAGE_TAG} -f image/Dockerfile ."

                docker.withRegistry("https://cargo.caicloudprivatetest.com", "cargo-private-admin") {
                    docker.image(IMAGE_TAG).push()
                }
                if (params.autoGitTag) {
                    echo "auto git tag: " + params.imageTag
                    withCredentials ([[$class: 'UsernamePasswordMultiBinding', credentialsId: 'caicloud-bot', usernameVariable: 'GIT_USERNAME', passwordVariable: 'GIT_PASSWORD']]){
                        sh("git config --global user.email \"info@caicloud.io\"")
                        sh("git tag -a $imageTag -m \"$tagDescribe\"")
                        sh("git push https://${GIT_USERNAME}:${GIT_PASSWORD}@github.com/caicloud/helm-registry $imageTag")
                   }
                } 
            }
        }

        stage('Deploy') {
            if (!params.deploy) {
                echo "skip deploy"
                return
            }
            def kubeconfig = "kubeconfig-${params.deployTarget}"
            withCredentials([[$class: 'FileBinding', credentialsId: kubeconfig, variable: 'SECRET_FILE']]) {
                sh("""
                    kubectl --kubeconfig=$SECRET_FILE --namespace default get deploy helm-registry-v0.1.0 -o yaml | sed 's/helm-registry:.*\$/helm-registry:${params.imageTag}/' | kubectl --kubeconfig=$SECRET_FILE --namespace default replace -f -
                """)
            }
        }
    }
}
