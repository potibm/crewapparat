group "default" {
  targets = ["crewapparat"]
}

target "crewapparat" {
  context    = "."
  dockerfile = "Dockerfile"
  platforms  = ["linux/amd64", "linux/arm64"]

  labels = {
    "org.opencontainers.image.url" = "https://github.com/potibm/crewapparat"
    "org.opencontainers.image.source" = "https://github.com/potibm/crewapparat"
    "org.opencontainers.image.documentation" = "https://github.com/potibm/crewapparat/tree/main/doc"
    "org.opencontainers.image.authors" = "potibm"
  }
  
  annotations = [
    "index,manifest:org.opencontainers.image.title=Crewapparat",
    "index,manifest:org.opencontainers.image.description=Crew management for demoparties.",
    "index,manifest:org.opencontainers.image.url=https://github.com/potibm/crewapparat",
    "index,manifest:org.opencontainers.image.source=https://github.com/potibm/crewapparat",
    "index,manifest:org.opencontainers.image.documentation=https://github.com/potibm/crewapparat/tree/main/doc",
    "index,manifest:org.opencontainers.image.licenses=MIT",
    "index,manifest:org.opencontainers.image.authors=potibm"
  ]
}
