package scheduler

import (
	"github.com/yanchenxu/Web-spider/middleware"
	"github.com/yanchenxu/Web-spider/itemProcessor"
	"github.com/yanchenxu/Web-spider/downloader"
	"github.com/yanchenxu/Web-spider/analyzer"
	"net/http"
	"errors"
	"strings"
	"regexp"
	"fmt"
)

func generateChannelManager(chnnellen uint) middleware.ChannelManager {
	return middleware.NewChannelManager(chnnellen)
}

func generatePageDownloaderPool(poolSize uint32, genClient GenHttpClient) (Downloader.PageDownloaderPool, error) {
	return Downloader.NewPageDownloaderPool(poolSize, Downloader.NewDownloder(genClient()))
}

func generateAnalyzerPool(poolSize uint32) (analyzer.AnalyzerPool, error) {
	return analyzer.NewAnalyzerPool(poolSize, analyzer.NewAnalyzer())
}

func generateItemPipeline(itemProcessors []ItemProcessor.ProcessItem) ItemProcessor.ItemPipeline {
	return ItemProcessor.NewItemPipeline(itemProcessors)
}

func generateCode(prefix string, id uint32) string {
	return fmt.Sprintf("%s-%d", prefix, id)
}

func parseCode(code string)[]string{
	result:=make([]string,2)
	var codePrefix string
	var id string
}
//匹配ip地址的正则
var regexpForIp = regexp.MustCompile(`((?:(?:25[0-5]|2[0-4]\d|[01]?\d?\d)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d?\d))`)

var regexpForDomains = []*regexp.Regexp{
	// *.xx or *.xxx.xx
	regexp.MustCompile(`\.(com|com\.\w{2})$`),
	regexp.MustCompile(`\.(gov|gov\.\w{2})$`),
	regexp.MustCompile(`\.(net|net\.\w{2})$`),
	regexp.MustCompile(`\.(org|org\.\w{2})$`),
	// *.xx
	regexp.MustCompile(`\.me$`),
	regexp.MustCompile(`\.biz$`),
	regexp.MustCompile(`\.info$`),
	regexp.MustCompile(`\.name$`),
	regexp.MustCompile(`\.mobi$`),
	regexp.MustCompile(`\.so$`),
	regexp.MustCompile(`\.asia$`),
	regexp.MustCompile(`\.tel$`),
	regexp.MustCompile(`\.tv$`),
	regexp.MustCompile(`\.cc$`),
	regexp.MustCompile(`\.co$`),
	regexp.MustCompile(`\.\w{2}$`),
}

func getPrimaryDomain(host string) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", errors.New("The host is empty!")
	}
	if regexpForIp.MatchString(host) {
		return host, nil
	}

	var surfixIndex int

	for _, re := range regexpForDomains {
		pos := re.FindStringIndex(host)
		if pos != nil {
			surfixIndex = pos[0]
			break
		}
	}

	if surfixIndex > 0 {
		var pdIndex int
		firstPart := host[:surfixIndex]
		index := strings.LastIndex(firstPart, ".")
		if index < 0 {
			pdIndex < 0
		} else {
			pdIndex = index + 1
		}
		return host[pdIndex:], nil
	} else {
		return "", errors.New("Unrecognized host!")
	}

}