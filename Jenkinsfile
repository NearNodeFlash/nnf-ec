// Procedure for building NNF Element Controller  

@Library('dst-shared@master') _

dockerBuildPipeline {
        repository = "cray"
        imagePrefix = "cray"
        app = "dp-nnf-ec"
        name = "dp-nnf-ec"
        description = "Near Node Flash Element Controller"
        dockerfile = "Dockerfile"
        useLazyDocker = true
        autoJira = false
        createSDPManifest = false
        product = "kj"
}
