package scheduler

import (
	"errors"
	"fmt"
	"github.com/yanchenxu/Web-spider/analyzer"
	"github.com/yanchenxu/Web-spider/base"
	"github.com/yanchenxu/Web-spider/downloader"
	"github.com/yanchenxu/Web-spider/itemProcessor"
	"github.com/yanchenxu/Web-spider/middleware"
	"regexp"
	"strings"
)

func generateChannelManager(chnnelArgs base.ChannelArgs) middleware.ChannelManager {
	return middleware.NewChannelManager(chnnelArgs)
}

func generatePageDownloaderPool(poolBaseArgs base.PoolBaseArgs, genClient GenHttpClient) (Downloader.PageDownloaderPool, error) {
	return Downloader.NewPageDownloaderPool(poolBaseArgs.PageDownloaderPoolSize(), func() Downloader.PageDownloader {
		return Downloader.NewDownloder(genClient())
	})
}

func generateAnalyzerPool(poolBaseArgs base.PoolBaseArgs) (analyzer.AnalyzerPool, error) {
	return analyzer.NewAnalyzerPool(poolBaseArgs.AnalyzerPoolSize(), analyzer.NewAnalyzer)
}

func generateItemPipeline(itemProcessors []ItemProcessor.ProcessItem) ItemProcessor.ItemPipeline {
	return ItemProcessor.NewItemPipeline(itemProcessors)
}

func generateCode(prefix string, id uint32) string {
	return fmt.Sprintf("%s-%d", prefix, id)
}

func parseCode(code string) []string {
	result := make([]string, 2)
	var codePrefix string
	var id string

	index := strings.Index(code, "-")
	if index > 0 {
		codePrefix = code[:index]
		id = code[index+1:]
	} else {
		codePrefix = code
	}
	result[0] = codePrefix
	result[1] = id
	return result
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
			pdIndex = 0
		} else {
			pdIndex = index + 1
		}
		return host[pdIndex:], nil
	} else {
		return "", errors.New("Unrecognized host!")
	}

}
