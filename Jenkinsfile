def IMAGE_TAG = "caicloud/helm-registry:ci-${env.BUILD_NUMBER}"

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
                    sh('''
                        set -e
                        cd ${WORKDIR}
                        make test
                    ''')
                }
            }

            stage("Build image and publish") {
                sh "docker build -t ${IMAGE_TAG} -f image/Dockerfile ."

                docker.withRegistry("https://cargo.caicloudprivatetest.com", "cargo-private-admin") {
                    docker.image(IMAGE_TAG).push()
                }
            }
        }

        stage('Deploy') {
            sh('''
                echo "skip deployment"
            ''')
        }
    }
}
