docker_image = "golang:1.12"
docker_volume = "muss_go_mods"

def run(cmd) {
  sh """
    docker run \
      -v ${docker_volume}:/go -v "\$WORKSPACE:/mnt" -w /mnt \
      -e GOPRIVATE -e GO111MODULE \
      --rm ${docker_image} ${cmd}
  """
}
def go(cmd) {
  run("go ${cmd} ./...")
}

pipeline {
  agent {
    label "bridge-docker"
  }
  options {
    ansiColor('xterm')
  }
  environment {
    GOPRIVATE = 'gerrit.instructure.com'
  }
  stages {
    stage('Build') {
      steps {
        sh """docker pull ${docker_image}"""
        go("build")
      }
    }

    stage('Tests') {
      parallel {
        stage('Test') {
          steps {
            go("test -v")
          }
        }
        stage('Vet') {
          steps {
            go("vet")
          }
        }
      }
    }
  }
  post {
    always {
      sh """
        docker volume rm ${docker_volume}
      """
    }
  }
}
