package main

import (
	"github.com/docker/distribution/reference"
	registryclient "github.com/flant/docker-registry-client/registry"
	"github.com/romana/rlog"
	"net/http"
	"strings"
)

// TODO данные для доступа к registry серверам нужно хранить в secret-ах.
// TODO по imageInfo.Registry брать данные и подключаться к нужному registry.
// Пока известно, что будет только registry.gitlab.company.com

var DockerRegistryInfo = map[string]map[string]string{
	"registry.gitlab.company.com": map[string]string{
		"url":      "https://registry.gitlab.company.com",
		"user":     "oauth2",
		"password": "qweqwe",
	},
	// minikube specific
	"localhost:5000": map[string]string{
		"url": "http://kube-registry.kube-system.svc.cluster.local:5000",
	},
}

//const DockerRegistryUrl = "https://registry.gitlab.company.com"
//const DockerRegistryUser = "oauth2"
//const DockerRegistryToken = ""

type DockerImageInfo struct {
	Registry   string
	Repository string
	Tag        string
}

func DockerRegistryGetImageId(image string) (string, error) {
	imageInfo, err := DockerParseImageName(image)
	if err != nil {
		rlog.Errorf("REGISTRY Problem parsing image %s: %v", image, err)
		return "", err
	}

	url := ""
	user := ""
	password := ""
	if info, has_info := DockerRegistryInfo[imageInfo.Registry]; has_info {
		url = info["url"]
		user = info["user"]
		password = info["password"]
	}

	// Установить соединение с registry
	registry := NewDockerRegistry(url, user, password)

	// Получить описание образа
	antiopaManifest, err := registry.ManifestV2(imageInfo.Repository, imageInfo.Tag)
	if err != nil {
		rlog.Errorf("REGISTRY cannot get manifest for %s:%s: %v", imageInfo.Repository, imageInfo.Tag, err)
		return "", err
	}

	imageID := antiopaManifest.Config.Digest.String()
	rlog.Debugf("REGISTRY id=%s for %s:%s", imageID, imageInfo.Repository, imageInfo.Tag)

	return imageID, nil
}

func DockerParseImageName(imageName string) (imageInfo DockerImageInfo, err error) {
	namedRef, err := reference.ParseNormalizedNamed(imageName)
	switch {
	case err != nil:
		return
	case reference.IsNameOnly(namedRef):
		// Если имя без тэга, то docker добавляет latest
		namedRef = reference.TagNameOnly(namedRef)
	}

	tag := ""
	if tagged, ok := namedRef.(reference.Tagged); ok {
		tag = tagged.Tag()
	}

	imageInfo = DockerImageInfo{
		Registry:   reference.Domain(namedRef),
		Repository: reference.Path(namedRef),
		Tag:        tag,
	}

	rlog.Debugf("REGISTRY image %s parsed to reg=%s repo=%s tag=%s", imageName, imageInfo.Registry, imageInfo.Repository, imageInfo.Tag)

	return
}

func RegistryClientLogCallback(format string, args ...interface{}) {
	rlog.Debugf(format, args...)
}

// NewDockerRegistry - ручной конструктор клиента, как рекомендовано в комментариях
// к registryclient.New.
// Этот конструктор не запускает registry.Ping и логирует события через rlog.
func NewDockerRegistry(registryUrl, username, password string) *registryclient.Registry {
	url := strings.TrimSuffix(registryUrl, "/")
	transport := http.DefaultTransport
	transport = registryclient.WrapTransport(transport, url, username, password)

	return &registryclient.Registry{
		URL: url,
		Client: &http.Client{
			Transport: transport,
		},
		Logf: RegistryClientLogCallback,
	}
}
