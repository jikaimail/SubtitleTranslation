// main
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gitote/chardet"
	"github.com/huichen/sego"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

type subpart struct {
	SPos  int    `json:"sPos"`
	STime string `json:"sTime"`
	SCSub string `json:"sCSub"`
	SSub  string `json:"sSub"`
}

type subInfo struct {
	DPos      int       `json:"dPos"`
	DCSub     string    `json:"dCSub"`
	DESub     string    `json:"dESub"`
	MNum      int       `json:"Num"`
	SplitInfo []subpart `json:"SplitInfo"`
}

type Subtitles struct {
	Subtitles []subInfo `json:"Subtitles"`
}

type charCode struct {
	Name string
	num  int
}

type charCodes []charCode

var (
	h            bool
	slang        string
	sstype       string
	infilepath   string
	trfilepath   string
	josnfilepath string
	pgfilepath   string
	nplinenum    int
)

func init() {
	flag.BoolVar(&h, "h", false, "this help")
	flag.StringVar(&slang, "lang", "chs", "this Language option")
	flag.StringVar(&infilepath, "infile", "", "enter the file name here. \n (Requires plain srt subtitle file)")
	flag.StringVar(&trfilepath, "trfile", "", "enter the translate file name here.")
	flag.StringVar(&josnfilepath, "jsfile", "", "enter the json file name here.")
	flag.StringVar(&pgfilepath, "pfile", "", "Add punctuation to the original subtitles.")
	flag.IntVar(&nplinenum, "npline", 6, "How many lines of subtitles are there without punctuation? ")
	flag.StringVar(&sstype, "stype", "b", "this Subtitle option")

	// 改变默认的 Usage，flag包中的Usage 其实是一个函数类型。这里是覆盖默认函数实现，具体见后面Usage部分的分析
	flag.Usage = l_usage

}

func l_usage() {

	if slang == "en" {
		fmt.Fprintf(os.Stderr,
			`Subtitle translation / version: TrSubtitle/0.5  
   by jikai Email:jikaimail@gmail.com
Usage: 
  This software is used to extract the original text content in the untitled SRT
original subtitles as the original text to be translated; The user translates 
the original to be translated into a translation by using a website or a human 
translation;The software then analyzes the translation and the original 
subtitles to generate final subtitles.
1)TrSubtitle -infile Original subtitle file name
  Generate the original text of the subtitle to be translated;
2)The generated original text to be translated is stored in a file after being 
turned over by the website or manually translated;Make sure the translation 
matches the line position and total number of lines in the original content.
3)TrSubtitle -infile Original subtitle file name -trfile translation file name
  Generate the required subtitle file.
4)TrSubtitle -jsfile json filename
  If you need to further adjust the subtitles, you can correct the subtitle 
content in the json file;The program regenerates the required subtitle file 
according to the adjusted json file.(only support this software json format)
Options:
-h : help
-lang : chs display Chinese help en display English help. Default chs.
-infile : Enter the name of the original subtitle file to be processed 
  (requires unformatted SRT subtitle file)
-trfile : Enter the name of the translation file.
-jsfile : Enter the json file name.
-stype : o Generate translated subtitle file 
         b Generate bilingual subtitle files  Default b.
-pfile : Add punctuation to the original subtitles(Europarl Corpus)
-npline : How many lines of subtitles are there without punctuation? default 6
latest version:【https://github.com/jikaimail/SubtitleTranslation/releases】
`)

	} else {
		fmt.Fprintf(os.Stderr,
			`机翻双语字幕辅助软件/版本 : TrSubtitle/0.5
                  by jikai  Email:jikaimail@gmail.com
   此软件用于将无格式的SRT原文字幕中的原文内容提取为待译原文；
使用者通过机翻网站或者人工翻译将待译原文翻译成为译文；
软件再将译文与原文字幕进行分析处理并生成最终字幕。
1) TrSubtitle -infile 原文字幕文件名
生成字幕原文待译文件；
2) 将生成的待译原文文件通过机翻网站或者人工翻译后存储在一个文件内；
确保译文与原文内容的行位置和总行数要匹配。
3) TrSubtitle -infile 原文字幕文件名 -trfile 译文文件名
生成所需的字幕文件。
4) TrSubtitle -jsfile json文件名
如果需要对字幕进一步调整，可在json文件内对字幕内容进行修正；
程序根据调整后的json文件，重新生成所需的字幕文件。(仅支持本软件json格式)   
参数选项:
-h : 帮助
-lang   : chs显示中文帮助 en显示英文帮助. 默认chs.
-infile : 输入要处理的原文字幕文件名(需要无格式的SRT字幕文件)
-trfile : 输入译文文件名.
-jsfile : 输入json文件名.
-stype  : o 仅生成译文字幕 b 生成双语字幕文件 默认b.  
-pfile  : 为原文字幕添加标点符号.(仅Europarl Corpus)
-npline : 多少行原文字幕无标点符号时提示？默认 6
最新版本：【https://github.com/jikaimail/SubtitleTranslation/releases】
`)

	}

}

func (p charCodes) Len() int { return len(p) }

func (p charCodes) Less(i, j int) bool {
	return p[i].num > p[j].num
}

func (p charCodes) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func checkError(e error) {
	if e != nil {
		panic(e)
	}

}

//convert  GBK to UTF-8
func DecodeGBK(s []byte) ([]byte, error) {

	I := bytes.NewReader(s)
	O := transform.NewReader(I, simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}

//convert UTF-8 to GBK
func EncodeGBK(s []byte) ([]byte, error) {
	I := bytes.NewReader(s)
	O := transform.NewReader(I, simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}

//convert BIG5 to UTF-8
func DecodeBig5(s []byte) ([]byte, error) {
	I := bytes.NewReader(s)
	O := transform.NewReader(I, traditionalchinese.Big5.NewDecoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}

//convert UTF-8 to BIG5
func EncodeBig5(s []byte) ([]byte, error) {
	I := bytes.NewReader(s)
	O := transform.NewReader(I, traditionalchinese.Big5.NewEncoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}
func del_file(filename string) bool {
	_, fErr := os.Stat(filename)
	if !os.IsNotExist(fErr) {
		os.Remove(filename)
		return true
	} else {
		return false
	}
}
func containEndSym(endSym string) bool {
	//英文符号 逗号，句号，问号，感叹号，冒号，分号和破折号
	str := [...]string{",", ".", "?", "!", ":", ";", "-"}
	bresult := false

	for i := range str {
		if strings.Compare(endSym, str[i]) == 0 {
			bresult = true
			break
		}

	}
	return bresult
}

func ContainSym(tsym string) bool {
	//中文符号 逗号，句号，引号，问号，感叹号，分号，括号
	str := [...]string{"，", "。", "”", "？", "！", "；", "）", ")"}

	bresult := false

	for i := range str {
		if strings.Compare(tsym, str[i]) == 0 {
			bresult = true
			break
		}

	}
	return bresult
}

func JsonGenSub() {

	_, lerr := os.Stat(josnfilepath)
	if os.IsNotExist(lerr) {
		if slang == "en" {
			fmt.Print("No json files found:" + josnfilepath)
			fmt.Print("-jsfile json filename" + "\n")
		} else {
			fmt.Print("未发现json文件:" + josnfilepath)
			fmt.Print("-jsfile json 文件名 " + "\n")
		}
		os.Exit(0)
	}

	var jSub Subtitles

	jsfile, _ := ioutil.ReadFile(josnfilepath)
	//去掉utf8 BOM标志
	jsfile = bytes.Replace(jsfile, []byte("\uFEFF"), []byte(""), 1)

	_ = json.Unmarshal([]byte(jsfile), &jSub.Subtitles)

	jschsfilename := ""

	if strings.HasSuffix(strings.ToLower(josnfilepath), ".json") {
		tempath := josnfilepath[0 : len(josnfilepath)-5]
		if strings.HasSuffix(strings.ToLower(tempath), ".srt") {
			jschsfilename = josnfilepath[0:len(tempath)-4] + ".chs.srt"
		} else {
			jschsfilename = josnfilepath + ".txt"
		}

	} else {
		jschsfilename = josnfilepath + ".txt"
	}

	del_file(jschsfilename)

	modifyfile, mErr := os.OpenFile(jschsfilename, os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	checkError(mErr)
	defer modifyfile.Close()

	jssubtext := ""

	//根据json 文件直接生成双语字幕
	//增加传播字幕行，让更多人受益。
	jsfirstext := "1" + "\n" +
		"00:00:00,000 --> 00:00:05,000" + "\n" +
		"{\\pos(200,210)}由机翻双语字幕辅助软件直接生成，此字幕仅用于研究学习" + "\n" +
		"url：github.com/jikaimail/SubtitleTranslation/" + "\n"
	_, werr := modifyfile.WriteString(jsfirstext)
	checkError(werr)

	for jspos := range jSub.Subtitles {

		for spos := range jSub.Subtitles[jspos].SplitInfo {
			if sstype == "b" {
				jssubtext = strconv.Itoa(jSub.Subtitles[jspos].SplitInfo[spos].SPos+1) + "\n" +
					jSub.Subtitles[jspos].SplitInfo[spos].STime + "\n" +
					jSub.Subtitles[jspos].SplitInfo[spos].SCSub + "\n" +
					jSub.Subtitles[jspos].SplitInfo[spos].SSub + "\n"
			} else {
				jssubtext = strconv.Itoa(jSub.Subtitles[jspos].SplitInfo[spos].SPos+1) + "\n" +
					jSub.Subtitles[jspos].SplitInfo[spos].STime + "\n" +
					jSub.Subtitles[jspos].SplitInfo[spos].SCSub + "\n"
			}
			_, werr := modifyfile.WriteString(jssubtext)
			checkError(werr)
		}
	}
	if slang == "en" {
		fmt.Println("Generate subtitle file from json file. ")
		fmt.Println("Please check the file: " + jschsfilename + " ." + "\n")
	} else {
		fmt.Println("由json文件生成字幕文件.")
		fmt.Println("请查看文件: " + jschsfilename + " ." + "\n")
	}
	os.Chmod(jschsfilename, 0644)
}

func oSubGentrText(inpath string) []subInfo {
	var insub []subInfo
	var CurSub subInfo
	var curpart subpart
	CurSub = subInfo{}

	var NewSub string
	preTime := ""
	BomLine := 1
	lineCn := 1
	NewCn := 1
	mNum := 0
	lend := false
	bnpline := false

	file, err := os.Open(inpath)
	checkError(err)
	defer file.Close()

	del_file(inpath + ".en.txt")

	enfile, enErr := os.OpenFile(inpath+".en.txt", os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	checkError(enErr)
	defer enfile.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		subText := scanner.Text()

		//匹配空行
		nilreg := regexp.MustCompile(`^$`)
		if nilreg.MatchString(subText) {
			continue
		}
		//判断行号
		linereg := regexp.MustCompile(`^[0-9]+$`)
		//去掉utf8 BOM标志
		if BomLine == 1 {
			subDigit := []rune(subText)
			for idigit := range subDigit {
				if subDigit[idigit] == '\uFEFF' {
					subText = strings.Replace(subText, "\uFEFF", "", 1)
				}
			}
			BomLine++
		}
		if linereg.MatchString(subText) {
			if strconv.Itoa(lineCn) == subText {
				if lineCn > 1 && lend {
					NewSub += "\n"
					//替换影响机器翻译质量的 - 空格 符号
					NewSub = strings.Replace(NewSub, "-", " ", -1)
					NewSub = strings.Replace(NewSub, "  ", " ", -1)
					NewSub = strings.Replace(NewSub, "  ", " ", -1)

					_, werr := enfile.WriteString(NewSub)
					checkError(werr)

					CurSub.DPos = NewCn
					CurSub.MNum = mNum

					if mNum >= nplinenum && len(pgfilepath) == 0 {
						if !bnpline {
							if slang == "en" {
								fmt.Println("The lack of punctuation will greatly affect the subtitle translation effect.")
							} else {
								fmt.Println("缺少标点符号将极大影响字幕翻译效果，建议人工添加标点符号！")
							}
							bnpline = true
						}
						fmt.Println("BeginPos：" + strconv.Itoa(CurSub.SplitInfo[0].SPos) + " - EndPos：" +
							strconv.Itoa(CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SPos) + "  Rows:" + strconv.Itoa(mNum))
					}
					CurSub.DESub = NewSub
					insub = append(insub, CurSub)
					NewCn++
					CurSub = subInfo{}
					NewSub = ""
					mNum = 0
				}

				//fmt.Println(subText)
				curpart = subpart{}
				curpart.SPos = lineCn
				lineCn++
				continue
			}
		}

		//判断时间轴
		timereg := regexp.MustCompile(`^\d*:\d*:\d*\d*:*,\d* --> \d*:\d*:\d*\d*:*,\d*$`)
		if timereg.MatchString(subText) {
			//fmt.Println(subText)
			curpart.STime = subText
			continue
		}
		//判断行尾
		reg := regexp.MustCompile(`([;\.\?!])\"*$`)
		if reg.MatchString(subText) {
			//fmt.Println("--" + subText)
			NewSub += subText + " "
			if (preTime == curpart.STime) && (BomLine >= 2) {
				CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SSub += " " + subText
				CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SSub = strings.Replace(CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SSub, "-", "", -1)
				CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SSub = strings.Replace(CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SSub, "  ", " ", -1)
				CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SSub = strings.Replace(CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SSub, "  ", " ", -1)

			} else {
				curpart.SSub = subText
				curpart.SSub = strings.Replace(curpart.SSub, "-", "", -1)
				curpart.SSub = strings.Replace(curpart.SSub, "  ", " ", -1)
				curpart.SSub = strings.Replace(curpart.SSub, "  ", " ", -1)
				CurSub.SplitInfo = append(CurSub.SplitInfo, curpart)
				mNum++
				preTime = curpart.STime
			}
			lend = true
		} else {
			NewSub += subText + " "

			if (preTime == curpart.STime) && (BomLine >= 2) {
				CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SSub += " " + subText
				CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SSub = strings.Replace(CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SSub, "-", "", -1)
				CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SSub = strings.Replace(CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SSub, "  ", " ", -1)
				CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SSub = strings.Replace(CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SSub, "  ", " ", -1)
			} else {
				curpart.SSub = subText
				curpart.SSub = strings.Replace(curpart.SSub, "-", "", -1)
				curpart.SSub = strings.Replace(curpart.SSub, "  ", " ", -1)
				curpart.SSub = strings.Replace(curpart.SSub, "  ", " ", -1)
				CurSub.SplitInfo = append(CurSub.SplitInfo, curpart)
				mNum++
			}
			preTime = curpart.STime
			lend = false
		}

	}

	NewSub += "\n"
	//替换影响机器翻译质量的 - 符号
	NewSub = strings.Replace(NewSub, "-", " ", -1)
	NewSub = strings.Replace(NewSub, "  ", " ", -1)
	NewSub = strings.Replace(NewSub, "  ", " ", -1)

	_, werr := enfile.WriteString(NewSub)
	checkError(werr)

	CurSub.DPos = NewCn
	CurSub.MNum = mNum
	CurSub.DESub = NewSub
	insub = append(insub, CurSub)
	CurSub = subInfo{}
	NewSub = ""
	mNum = 0

	os.Chmod(inpath+".en.txt", 0644)
	return insub
}

func DetectFCharset(dfile string, ccF charCodes) charCodes {
	trCode, chsErr := os.Open(dfile)
	checkError(chsErr)
	defer trCode.Close()

	codeScanner := bufio.NewScanner(trCode)
	alltext := ""
	for codeScanner.Scan() {
		alltext += codeScanner.Text() + "\n "
	}
	ltext := []byte(alltext)
	textDetector := chardet.NewTextDetector()
	Result, _ := textDetector.DetectBest(ltext)
	for a := 0; a < len(ccF)-1; a++ {
		if strings.Compare(ccF[a].Name, Result.Charset) == 0 {
			ccF[a].num = ccF[a].num + 1
			//fmt.Println(tCA[a].Name + strconv.Itoa(tCA[a].num))
			break
		}
	}
	sort.Sort(ccF)
	return ccF
}

func chstolastSub(chsallsub []subInfo) []subInfo {
	trchsfilename := ""
	chsfile, chsErr := os.Open(trfilepath)
	checkError(chsErr)
	defer chsfile.Close()

	if strings.HasSuffix(strings.ToLower(infilepath), ".srt") {
		trchsfilename = infilepath[0:len(infilepath)-4] + ".chs.srt"
	} else {
		trchsfilename = infilepath + ".txt"
	}

	del_file(trchsfilename)

	subfile, enErr := os.OpenFile(trchsfilename, os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	checkError(enErr)
	defer subfile.Close()
	// 载入词典
	var segmenter sego.Segmenter
	segmenter.LoadDictionary("dictionary.txt")

	//确定翻译文件字符集
	var tCA charCodes
	tCA = []charCode{{"UTF8", 0}, {"GB18030", 0}, {"BIG5", 0}}

	tCA = DetectFCharset(trfilepath, tCA)
	if slang == "en" {
		fmt.Println("Determine the character set：" + tCA[0].Name + "\n")

	} else {
		fmt.Println("确定翻译文件的字符集为：" + tCA[0].Name + "\n")
	}

	//开始合并翻译文件
	//增加传播字幕行，让更多人受益。
	jsfirstxt := "1" + "\n" +
		"00:00:00,000 --> 00:00:05,000" + "\n" +
		"{\\pos(200,210)}由机翻双语字幕辅助软件直接生成，此字幕仅用于研究学习" + "\n" +
		"url：github.com/jikaimail/SubtitleTranslation/" + "\n"
	_, suberr := subfile.WriteString(jsfirstxt)
	checkError(suberr)

	chsScanner := bufio.NewScanner(chsfile)
	lCount := 0

	for chsScanner.Scan() {
		//去掉utf8 BOM标志
		bomtext := ""
		if lCount == 0 {
			bomtext = strings.Replace(chsScanner.Text(), "\uFEFF", "", 1)

		} else {
			bomtext = chsScanner.Text()
		}
		temptext := []byte(bomtext)
		switch tCA[0].Name {
		case "GB18030":
			stem, _ := DecodeGBK(temptext)
			chsallsub[lCount].DCSub = string(stem)
		case "UTF8":
			chsallsub[lCount].DCSub = chsScanner.Text()
		case "BIG5":
			stem, _ := DecodeBig5(temptext)
			chsallsub[lCount].DCSub = string(stem)
		default:
			chsallsub[lCount].DCSub = chsScanner.Text()
		}
		var subtext string

		//将每句翻译，切分为若干行
		if chsallsub[lCount].MNum > 1 {
			lastEnSub := chsallsub[lCount].DESub
			//替换（,）为（，）,同时处理数字的，逗号问题。
			lastSub := strings.Replace(chsallsub[lCount].DCSub, ",", "，", -1)
			lastSub = strings.Replace(lastSub, "  ", " ", -1)
			lastSub = strings.Replace(lastSub, "  ", " ", -1)
			lastSub = strings.Replace(lastSub, "  ", " ", -1)
			lastSub = strings.Replace(lastSub, "  ", " ", -1)

			lastEnSub = strings.Replace(lastEnSub, "  ", " ", -1)
			lastEnSub = strings.Replace(lastEnSub, "  ", " ", -1)
			lastEnSub = strings.Replace(lastEnSub, "  ", " ", -1)
			lastEnSub = strings.Replace(lastEnSub, "  ", " ", -1)

			regdig := regexp.MustCompile(`[0-9]+[，][0-9]+`)
			regdigf := func(s string) string {
				return strings.Replace(s, "，", ",", -1)
			}
			digfstr := regdig.ReplaceAllStringFunc(lastSub, regdigf)
			lastSub = digfstr

			regSplit := regexp.MustCompile(`,$`)
			preSplit := true
			var enlen, chsLen int

			for i := range chsallsub[lCount].SplitInfo {
				var subchs string
				subchs = ""
				//fmt.Println(strconv.Itoa(i) + ":" + strconv.Itoa(chsallsub[lCount].MNum))

				//当仅一行或多行时的最后一行 则直接赋值
				if i == chsallsub[lCount].MNum-1 {
					subchs = lastSub
				} else {
					//切分行数大于1时
					bsplit := regSplit.MatchString(chsallsub[lCount].SplitInfo[i].SSub)
					sEn := strings.Split(chsallsub[lCount].SplitInfo[i].SSub, ",")
					sChs := strings.Split(lastSub, "，")

					//有逗号结尾分隔符切分
					if (len(sChs) >= len(sEn)) && bsplit && preSplit {
						for j := range sEn {
							juNum := len(lastSub) / chsallsub[lCount].MNum
							if j == len(sEn)-1 {
								//处理译文多出一个逗号的特殊情况

								if (len(sChs) > len(sEn)) &&
									(!(len(sChs) == chsallsub[lCount].MNum-1)) &&
									(len(subchs) < (juNum - juNum/3)) {
									nextNum := len(subchs + sChs[j])
									if nextNum < (juNum+juNum/3) &&
										(j < len(sChs)-1) {
										subchs += sChs[j]
										subchs += "，"
									}
								}
								break
							} else {
								if len(subchs+sChs[j]) > (juNum+juNum/3) &&
									len(subchs) > 0 {
									break
								}
							}
							subchs += sChs[j]
							subchs += "，"
						}
						preSplit = true
					} else {
						//无逗号结尾分隔符切分

						//获取剩余英文总长度
						textten := []byte(lastEnSub)
						segments2 := segmenter.Segment(textten)
						strtxt2 := sego.SegmentsToSlice(segments2, false)
						iennum := 0
						for ic := range strtxt2 {
							if !(strings.Compare(strtxt2[ic], " ") == 0) {
								iennum += 1
							}
						}
						enlen = iennum

						//获取当前行英文长度
						textten1 := []byte(chsallsub[lCount].SplitInfo[i].SSub)
						segments3 := segmenter.Segment(textten1)
						strtxt1 := sego.SegmentsToSlice(segments3, false)
						lennum := 0
						for ib := range strtxt1 {
							if !(strings.Compare(strtxt1[ib], " ") == 0) {
								lennum += 1
							}
						}
						var linlen float64
						linlen = float64(lennum)

						//获取剩余中文的长度
						text1 := []byte(lastSub)
						segments1 := segmenter.Segment(text1)
						strtxt := sego.SegmentsToSlice(segments1, false)
						chsnum := 0
						for ia := range strtxt {
							if !(strings.Compare(strtxt[ia], " ") == 0) {
								chsnum += 1
							}
						}
						chsLen = chsnum

						preSplit = false
						text := []byte(lastSub)
						segments := segmenter.Segment(text)
						CText := sego.SegmentsToSlice(segments, false)
						var nextpos, avgLen float64
						avgLen = float64(linlen) / float64(enlen) * float64(chsLen)
						nextpos = 0
						lsub := ""
						presub := ""

						for k := range CText {
							lsub = ""
							avgline := float64(avgLen / 3)
							if ((nextpos - avgLen) >= 0) && (!ContainSym(CText[k])) {
								var lpos float64
								lpos = 0
								for j := k; j < len(CText)-1; j++ {
									if (lpos <= avgline) && (j < len(CText)-2) {
										if ContainSym(CText[j]) {
											lsub += CText[j]
											subchs += lsub
											lsub = ""
											break
										}
										lsub += CText[j]
										if !(strings.Compare(CText[j], " ") == 0) {
											lpos += 1
										}
										continue
									} else {
										if len(presub) > 0 {
											subchs = presub
										} else {
											lsub = ""
										}
										break
									}
								}
								break
							}

							brnum := (len(chsallsub[lCount].SplitInfo) - i - 1) * 2
							if k < (len(CText) - brnum) {
								subchs += CText[k]
							} else {
								break
							}

							if !(strings.Compare(CText[k], " ") == 0) {
								nextpos += 1
							}
							if ContainSym(CText[k]) && ((avgLen - nextpos) <= avgline) {
								presub = subchs
							}
						}
					}
				}
				chsallsub[lCount].SplitInfo[i].SCSub = subchs
				lastEnSub = lastEnSub[len(chsallsub[lCount].SplitInfo[i].SSub):len(lastEnSub)]
				lastSub = lastSub[len(subchs):len(lastSub)]
				if sstype == "b" {
					subtext = strconv.Itoa(chsallsub[lCount].SplitInfo[i].SPos+1) + "\n" +
						chsallsub[lCount].SplitInfo[i].STime + "\n" +
						chsallsub[lCount].SplitInfo[i].SCSub + "\n" +
						chsallsub[lCount].SplitInfo[i].SSub + "\n"
				} else {
					subtext = strconv.Itoa(chsallsub[lCount].SplitInfo[i].SPos+1) + "\n" +
						chsallsub[lCount].SplitInfo[i].STime + "\n" +
						chsallsub[lCount].SplitInfo[i].SCSub + "\n"
				}
				//fmt.Println(subtext)
				//fmt.Println("POS:" + strconv.Itoa(chsallsub[lCount].DPos))
				_, werr := subfile.WriteString(subtext)
				checkError(werr)
			}

		} else {
			if chsallsub[lCount].MNum == 1 {
				chsallsub[lCount].SplitInfo[0].SCSub = chsallsub[lCount].DCSub
				if sstype == "b" {
					subtext = strconv.Itoa(chsallsub[lCount].SplitInfo[0].SPos+1) + "\n" +
						chsallsub[lCount].SplitInfo[0].STime + "\n" +
						chsallsub[lCount].DCSub + "\n" +
						chsallsub[lCount].SplitInfo[0].SSub + "\n"
				} else {
					subtext = strconv.Itoa(chsallsub[lCount].SplitInfo[0].SPos+1) + "\n" +
						chsallsub[lCount].SplitInfo[0].STime + "\n" +
						chsallsub[lCount].DCSub + "\n"
				}
				//fmt.Println(subtext)
				//fmt.Println("POS:" + strconv.Itoa(chsallsub[lCount].DPos))
				_, werr := subfile.WriteString(subtext)
				checkError(werr)

			}
		}
		lCount++
	}
	if slang == "en" {
		fmt.Println("A subtitle file has been generated .")
		fmt.Println("Please check the file: " + trchsfilename + " ." + "\n")
	} else {
		fmt.Println("生成所需的字幕文件.")
		fmt.Println("请查看文件: " + trchsfilename + " ." + "\n")
	}
	os.Chmod(trchsfilename, 0644)
	return chsallsub
}

func oSubAddPunctuator(oSubinfo []subInfo) {
	var segmenter sego.Segmenter
	segmenter.LoadDictionary("dictionary.txt")
	del_file(pgfilepath + ".en.srt")

	pgfile, enErr := os.OpenFile(pgfilepath+".en.srt", os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	checkError(enErr)
	defer pgfile.Close()
	subtext := ""

	if slang == "en" {
		fmt.Println("Punctuation is being accessed at http://bark.phon.ioc.ee/punctuator." + "\n")
	} else {
		fmt.Println("正在访问http://bark.phon.ioc.ee/punctuator获取标点符号。" + "\n")
	}

	for ia := range oSubinfo {

		resp, _ := http.PostForm("http://bark.phon.ioc.ee/punctuator",
			url.Values{"text": {oSubinfo[ia].DESub}})
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)

		segments1 := segmenter.Segment(body)
		pgText := sego.SegmentsToSlice(segments1, false)

		lcn := 0

		for ib := range oSubinfo[ia].SplitInfo {
			ensubtext := []byte(oSubinfo[ia].SplitInfo[ib].SSub)
			segments1 := segmenter.Segment(ensubtext)
			lsubText := sego.SegmentsToSlice(segments1, false)
			lastText := ""
			for ic := range lsubText {

				if lsubText[ic] == " " {
					continue
				}
				if lsubText[ic] == "'" {
					//lastText = lastText[0 : len(lastText)-1]
					lastText = strings.TrimRight(lastText, " ")
					lastText += lsubText[ic]
				} else {
					lastText += lsubText[ic] + " "
				}
				for id := lcn; id < len(pgText)-1; id++ {
					lcn = id

					if pgText[id] == " " ||
						pgText[id] == lsubText[ic] {
						continue
					}

					if containEndSym(pgText[id]) {
						lastText = strings.TrimRight(lastText, " ")
						lastText += pgText[id] + " "
					} else {
						break
					}
				}
			}
			lastText = strings.TrimRight(lastText, " ")
			lastText = strings.TrimRight(lastText, " ")

			subtext = strconv.Itoa(oSubinfo[ia].SplitInfo[ib].SPos) + "\n" +
				oSubinfo[ia].SplitInfo[ib].STime + "\n" +
				lastText + "\n"
			_, werr := pgfile.WriteString(subtext)
			checkError(werr)
		}
	}
	os.Chmod(pgfilepath+".en.srt", 0644)
	if slang == "en" {
		fmt.Println("Generate a subtitle file with punctuation added .")
		fmt.Println("Please check the file: " + pgfilepath + ".en.srt" + " ." + "\n")
	} else {
		fmt.Println("生成带添加标点符号的字幕文件.")
		fmt.Println("请查看文件: " + pgfilepath + ".en.srt" + " ." + "\n")
	}
}

func main() {
	flag.Parse()

	if h || (infilepath == "" && josnfilepath == "" && len(pgfilepath) == 0) {
		flag.Usage()
		os.Exit(0)
	}
	var allsub []subInfo

	//为原字幕文件添加标点符号
	if len(pgfilepath) > 0 {
		allsub = oSubGentrText(pgfilepath)
		del_file(pgfilepath + ".en.txt")
		oSubAddPunctuator(allsub)
		os.Exit(0)
	}

	//由辅助json文件直接生成双语字幕
	if len(josnfilepath) > 0 {
		JsonGenSub()
		os.Exit(0)
	}

	//转换原文字幕为待翻译文件
	_, lerr := os.Stat(infilepath)
	if os.IsNotExist(lerr) {
		if slang == "en" {
			fmt.Print("No subtitle files found:" + infilepath + "\n")
			fmt.Print("-infile filename (Requires plain srt subtitle file)" + "\n")
		} else {
			fmt.Print("未发现字幕文件:" + infilepath + "\n")
			fmt.Print("-infile 字幕文件名 (需要无格式的srt字幕文件)" + "\n")
		}
		os.Exit(0)
	}

	allsub = oSubGentrText(infilepath)

	if len(trfilepath) == 0 {
		if slang == "en" {
			fmt.Println("Please translate the file [" + infilepath + ".en.txt]  ")
			fmt.Println("Translate URLs: https://translate.google.com/")
			fmt.Println("             or https://cn.bing.com/Translator")
			fmt.Println("             or https://fanyi.baidu.com")
			fmt.Println("Chrome can drag and drop files directly onto the above website pages,  ")
			fmt.Println("  and Google Translate can generate translations directly.")
			fmt.Println("Note: Make sure the translated content matches the line location ")
			fmt.Println("      and total number of rows of the original content." + "\n")
		} else {
			fmt.Println("请翻译此文件 [" + infilepath + ".en.txt]  ")
			fmt.Println("可选用以下网址进行翻译： ")
			fmt.Println(" URLs: https://translate.google.com")
			fmt.Println("    or https://cn.bing.com/Translator")
			fmt.Println("    or https://fanyi.baidu.com")
			fmt.Println("Chrome可将文件直接拖拽到以上网站页面，谷歌翻译即可生成翻译内容。")
			fmt.Println("注意事项：确保翻译内容与原内容的行位置和总行数要匹配。" + "\n")
		}

		//生成辅助json文件
		del_file(infilepath + ".json")
		jfilein, _ := json.MarshalIndent(allsub, "", "\t")
		_ = ioutil.WriteFile(infilepath+".json", jfilein, 0644)

		os.Exit(0)
	}

	//处理并合并翻译文件
	_, eErr := os.Stat(trfilepath)
	if !os.IsNotExist(eErr) {

		allsub = chstolastSub(allsub)
	} else {
		if slang == "en" {
			fmt.Println("The translated subtitle file was not found." + "\n")
			fmt.Println("Please check if the file path and file name are correct." + "\n")
			fmt.Println("TrSubtitle -h Get help." + "\n")
		} else {
			fmt.Println("未发现已翻译的字幕文件，请核对文件路径及文件名是否正确。" + "\n")
			fmt.Println("TrSubtitle -h 获取帮助。" + "\n")
		}
	}

	//生成辅助json文件
	del_file(infilepath + ".json")
	jlastfile, _ := json.MarshalIndent(allsub, "", "\t")
	_ = ioutil.WriteFile(infilepath+".json", jlastfile, 0644)

}
