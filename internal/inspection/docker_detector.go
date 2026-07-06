package inspection

var dockerComposeFiles = []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}

type dockerDetector struct{}

// NewDockerDetector returns a Detector for containerized projects
// (Dockerfile and/or a Compose file). It says nothing about the language or
// framework running inside the container — that is left to the other
// detectors, which can coexist with this one on the same project.
func NewDockerDetector() Detector {
	return &dockerDetector{}
}

func (d *dockerDetector) Name() string { return "docker" }

func (d *dockerDetector) Detect(in DetectionInput) (Detection, bool) {
	hasDockerfile := in.Has("Dockerfile")

	var composeFile string
	for _, candidate := range dockerComposeFiles {
		if in.Has(candidate) {
			composeFile = candidate
			break
		}
	}

	if !hasDockerfile && composeFile == "" {
		return Detection{}, false
	}

	detection := Detection{
		Type:       ProjectTypeDocker,
		Confidence: ConfidenceHigh,
	}

	if hasDockerfile {
		detection.MatchedFiles = append(detection.MatchedFiles, "Dockerfile")
	}
	if composeFile != "" {
		detection.MatchedFiles = append(detection.MatchedFiles, composeFile)
	}

	switch {
	case composeFile != "":
		detection.PackageManager = "docker compose"
		detection.Suggested.Build = "docker compose build"
		detection.Suggested.Start = "docker compose up"
	case hasDockerfile:
		detection.PackageManager = "docker"
		detection.Suggested.Build = "docker build ."
		detection.Notes = append(detection.Notes,
			"no compose file found; a start command was not suggested because it needs an image name and port mapping")
	}

	return detection, true
}
