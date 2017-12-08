def registryImage = "caicloud/helm-registry:${params.imageTag}"



podTemplate(
   cloud: 'dev-cluster',
   namespace: 'kube-system',
   // change the label to your component name.
   label: 'helm-registry',
   containers: [
       // a Jenkins agent (FKA "slave") using JNLP to establish connection.
       containerTemplate(
           name: 'jnlp',
           // alwaysPullImage: true,
           image: 'cargo.caicloudprivatetest.com/caicloud/jenkins/jnlp-slave:3.14-1-alpine',
           command: '',
           args: '${computer.jnlpmac} ${computer.name}',
       ),
       // docker in docker
       containerTemplate(
           name: 'dind',
           image: 'cargo.caicloudprivatetest.com/caicloud/docker:17.09-dind',
           ttyEnabled: true,
           command: '',
           args: '--host=unix:///home/jenkins/docker.sock',
           privileged: true,
       ),
       // golang with docker client and tools
       containerTemplate(
           name: 'golang',
           image: 'cargo.caicloudprivatetest.com/caicloud/golang-docker:1.9-17.09',
           ttyEnabled: true,
           command: '',
           args: '',
           envVars: [
               containerEnvVar(key: 'DOCKER_HOST', value: 'unix:///home/jenkins/docker.sock'),
               // Change the environment variable WORKDIR as needed.
               containerEnvVar(key: 'WORKDIR', value: '/go/src/github.com/caicloud/helm-registry')
           ],
       )
   ]
) {
   // Change the node name as the podTemplate label you set.
   node('helm-registry') {
       stage('Checkout') {
          checkout scm
       }
       // Change the container name as the container you use for compiling.
       container('golang') {
           ansiColor('xterm') {
               // You can define the stage as you need.
               stage("Complie") {
                   sh('''
                       set -e
                       mkdir -p $(dirname ${WORKDIR})
                       rm -rf ${WORKDIR}
                       ln -sfv $(pwd) ${WORKDIR}


                       cd ${WORKDIR}
                      
                       make build
                   ''')
               }



               stage('Unit test') {
                   sh('''
                       set -e
                       cd ${WORKDIR}


                       make test
                   ''')
               }



               stage('Build and push image') {
                   sh('''
                       set -e
                       cd ${WORKDIR}
                   ''')



                   sh("docker build -t ${registryImage} -f build/registry/Dockerfile .")



                   // Whether publish the images is controlled by the params.
                   if (params.publish) {
                       docker.withRegistry("https://cargo.caicloudprivatetest.com", "cargo-private-admin") {
                           docker.image(registryImage).push()
                       }
                   }
               }
           }
       }
   }
}
