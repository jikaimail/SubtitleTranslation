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
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
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
	infilepath   string
	trfilepath   string
	josnfilepath string
)

func init() {
	flag.BoolVar(&h, "h", false, "this help")
	flag.StringVar(&slang, "lang", "chs", "this help")
	flag.StringVar(&infilepath, "infile", "", "enter the file name here. \n (Requires plain srt subtitle file)")
	flag.StringVar(&trfilepath, "trfile", "", "enter the translate file name here.")
	flag.StringVar(&josnfilepath, "jsfile", "", "enter the json file name here.")

	// 改变默认的 Usage，flag包中的Usage 其实是一个函数类型。这里是覆盖默认函数实现，具体见后面Usage部分的分析
	flag.Usage = l_usage

}

func l_usage() {

	if slang == "en" {
		fmt.Fprintf(os.Stderr,
			`Subtitle translation / version: TrSubtitle/0.2  
   by jikai Email:jikaimail@gmail.com
Usage: 
  Many videos have no Chinese subtitles for some reason; 
the software processes the original subtitles.
Improve the accuracy of the Chinese subtitles. 
It can also be used as a simple auxiliary tool for Chinese subtitle translation.
 1) TrSubtitle -infile subtitle file name
     Generate English documents to be translated; 
 2) The generated English file to be translated is manually translated 
and stored in a file; Make sure the translated content matches the line location 
and total number of rows of the original content.
 3) TrSubtitle -infile subtitle file name -trfile translated file name
     Generate the final bilingual subtitle file.
 4) TrSubtitle -jsfile json filename
     If you need to further adjust the subtitles, 
     you can correct the subtitle content in the json file;
     The program regenerates the bilingual subtitle file 
     according to the adjusted json file.     
Options:
  -h          : this help
  -lang       : chs display Chinese help, 
                en display English help default chs.
  -infile     : enter the file name here. 
               (Requires plain srt subtitle file)
  -trfile     : enter the translate file name here. 
  -jsfile     : enter the json file name here. 
latest version:
 【https://github.com/jikaimail/SubtitleTranslation/releases】
`)

	} else {
		fmt.Fprintf(os.Stderr,
			`机翻中文字幕辅助软件 / 版本: TrSubtitle/0.2
   by jikai  Email:jikaimail@gmail.com
使用方法：
   很多视频由于某种原因，没有中文字幕；本软件通过对原文字幕进行处理，
提高机翻中文字幕的准确性。 也可作为中文字幕翻译的简单辅助工具。
 1) TrSubtitle -infile 字幕文件名  
    生成待翻译的英文文件；
 2) 将生成的待翻译的英文文件，人工翻译并存储在一个文件内；
确保翻译内容与原内容的行位置和总行数要匹配。 
 3) TrSubtitle -infile 字幕文件名  -trfile 已翻译的文件名  
    生成最终双语字幕文件。
 4) TrSubtitle -jsfile json文件名
    如果需要对字幕进一步调整，可在json文件内对字幕内容进行修正；
    程序根据调整后的json文件，重新生成双语字幕文件。
    (仅支持本软件json格式)   
参数选项:
  -h          : 帮助
  -lang       : chs显示中文帮助 en显示英文帮助. 默认chs.
  -infile     : 输入要处理的字幕文件名. 
               (需要无格式的srt字幕文件)
  -trfile     : 输入已翻译文件名. 
  -jsfile     : 输入json文件名.    
最新版本：
  【https://github.com/jikaimail/SubtitleTranslation/releases】
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

func ContainSym(tsym string) bool {
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

func main() {
	flag.Parse()

	if h || (infilepath == "" && josnfilepath == "") {
		flag.Usage()
		os.Exit(0)
	}

	//由json文件直接生成双语字幕
	if len(josnfilepath) > 0 {
		_, lerr := os.Stat(josnfilepath)
		if os.IsNotExist(lerr) {
			if slang == "en" {
				fmt.Print("No json files found:" + josnfilepath + "\n")
				fmt.Print("-jsfile json filename" + "\n")
			} else {
				fmt.Print("未发现json文件:" + josnfilepath + "\n")
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
				jschsfilename = infilepath[0:len(tempath)-4] + ".chs.srt"
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
		for jspos := range jSub.Subtitles {

			for spos := range jSub.Subtitles[jspos].SplitInfo {

				jssubtext = strconv.Itoa(jSub.Subtitles[jspos].SplitInfo[spos].SPos) + "\n" +
					jSub.Subtitles[jspos].SplitInfo[spos].STime + "\n" +
					jSub.Subtitles[jspos].SplitInfo[spos].SCSub + "\n" +
					jSub.Subtitles[jspos].SplitInfo[spos].SSub + "\n"
				_, werr := modifyfile.WriteString(jssubtext)
				checkError(werr)
			}
		}

		if slang == "en" {
			fmt.Println("Generate bilingual subtitle files from json files. " + "\n")
			fmt.Println("Please check the file: " + jschsfilename + " ." + "\n")
		} else {
			fmt.Println("由json文件生成双语字幕文件." + "\n")
			fmt.Println("请查看文件: " + jschsfilename + " ." + "\n")
		}
		os.Chmod(jschsfilename, 0644)
		os.Exit(0)
	}

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


	file, err := os.Open(infilepath)
	checkError(err)
	defer file.Close()

	del_file(infilepath + ".en.txt")

	enfile, enErr := os.OpenFile(infilepath+".en.txt", os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	checkError(enErr)
	defer enfile.Close()

	lineCn := 1
	NewCn := 1
	lEnd := 1
	mNum := 0

	var NewSub string
	var CurSub subInfo
	var curpart subpart
	var allsub []subInfo
    preTime := ""
    BomLine := 1
	CurSub = subInfo{}

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
				if   subDigit[idigit] == '\uFEFF' {
					subText = strings.Replace(subText, "\uFEFF", "",1)
				}
			}
			BomLine ++
		}


		if linereg.MatchString(subText)  {
			if strconv.Itoa(lineCn) == subText {
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

			if lEnd == 1 {
				NewSub += subText
			} else {
				NewSub += subText
			}

			if (preTime == curpart.STime) && (BomLine >= 2) {
				CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SSub += subText
			} else {
				curpart.SSub = subText
				CurSub.SplitInfo = append(CurSub.SplitInfo, curpart)
				mNum++
			}

			NewSub += "\n"
			//替换影响机器翻译质量的 - 符号
			NewSub = strings.Replace(NewSub, "-", " ", -1)
			_, werr := enfile.WriteString(NewSub)
			checkError(werr)

			CurSub.DPos = NewCn
			CurSub.MNum = mNum
			CurSub.DESub = NewSub
			allsub = append(allsub, CurSub)
			CurSub = subInfo{}
			NewSub = ""
			mNum = 0
			NewCn++
		} else {
			NewSub += subText + " "

			if (preTime == curpart.STime) && (BomLine >= 2) {
				CurSub.SplitInfo[len(CurSub.SplitInfo)-1].SSub += subText
			} else {
				curpart.SSub = subText
				CurSub.SplitInfo = append(CurSub.SplitInfo, curpart)
				mNum++
			}

			preTime = curpart.STime
			lEnd = 0

		}

	}
	os.Chmod(infilepath+".en.txt", 0644)

	if len(trfilepath) == 0 {
		if slang == "en" {
			fmt.Println("Please translate the file [" + infilepath + ".en.txt]  " + "\n")
			fmt.Println("Translate URLs: https://translate.google.com/" + "\n")
			fmt.Println("             or https://cn.bing.com/Translator" + "\n")
			fmt.Println("             or https://fanyi.baidu.com" + "\n")
			fmt.Println("Chrome can drag and drop files directly onto the above website pages,  " + "\n")
			fmt.Println("  and Google Translate can generate translations directly." + "\n")
			fmt.Println("Note: Make sure the translated content matches the line location " + "\n")
			fmt.Println("      and total number of rows of the original content." + "\n")
		} else {
			fmt.Println("请翻译此文件 [" + infilepath + ".en.txt]  " + "\n")
			fmt.Println("可选用以下网址进行翻译： " + "\n")
			fmt.Println(" URLs: https://translate.google.com" + "\n")
			fmt.Println("    or https://cn.bing.com/Translator" + "\n")
			fmt.Println("    or https://fanyi.baidu.com" + "\n")
			fmt.Println("Chrome可将文件直接拖拽到以上网站页面，谷歌翻译即可生成翻译内容。" + "\n")
			fmt.Println("注意事项：确保翻译内容与原内容的行位置和总行数要匹配。" + "\n")
		}
		os.Exit(0)
	}

	//处理并合并翻译文件
	trchsfilename := ""
	_, eErr := os.Stat(trfilepath)
	if !os.IsNotExist(eErr) {
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
		trCode, chsErr := os.Open(trfilepath)
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
		for a := 0; a < len(tCA)-1; a++ {
			if strings.Compare(tCA[a].Name, Result.Charset) == 0 {
				tCA[a].num = tCA[a].num + 1
				//fmt.Println(tCA[a].Name + strconv.Itoa(tCA[a].num))
				break
			}
		}
		sort.Sort(tCA)
		if slang == "en" {
			fmt.Println("Determine the character set：" + tCA[0].Name + "\n")

		} else {
			fmt.Println("确定翻译文件的字符集为：" + tCA[0].Name + "\n")
		}

		//开始合并翻译文件
		chsScanner := bufio.NewScanner(chsfile)
		lCount := 0

		for chsScanner.Scan() {
			//去掉utf8 BOM标志
			bomtext := ""
			if lCount == 0 {
				bomtext = strings.Replace(chsScanner.Text(), "\uFEFF", "",1)

			} else {
				bomtext = chsScanner.Text()
			}
			temptext := []byte(bomtext)
			switch tCA[0].Name {
			case "GB18030":
				stem, _ := DecodeGBK(temptext)
				allsub[lCount].DCSub = string(stem)
			case "UTF8":
				allsub[lCount].DCSub = chsScanner.Text()
			case "BIG5":
				stem, _ := DecodeBig5(temptext)
				allsub[lCount].DCSub = string(stem)
			default:
				allsub[lCount].DCSub = chsScanner.Text()
			}
			var subtext string

			//将每句翻译，切分为若干行
			if allsub[lCount].MNum > 1 {
				lastEnSub := allsub[lCount].DESub
				//替换（,）为（，）,同时处理数字的，逗号问题。
				lastSub := strings.Replace(allsub[lCount].DCSub, ",", "，", -1)
				lastSub = strings.Replace(lastSub, "  ", " ", -1)
				lastSub = strings.Replace(lastSub, "  ", " ", -1)
				lastSub = strings.Replace(lastSub, "  ", " ", -1)
				lastSub = strings.Replace(lastSub, "  ", " ", -1)

				regdig := regexp.MustCompile(`[0-9]+[，][0-9]+`)
				regdigf := func(s string) string {
					 return strings.Replace(s,"，",",",-1)
				}
				digfstr :=  regdig.ReplaceAllStringFunc (lastSub, regdigf)
				lastSub =  digfstr


				regSplit := regexp.MustCompile(`,$`)
				preSplit := true
				var enlen, chsLen int

				for i := range allsub[lCount].SplitInfo {
					var subchs string
					subchs = ""
					//fmt.Println(strconv.Itoa(i) + ":" + strconv.Itoa(allsub[lCount].MNum))

					//当仅一行或多行时的最后一行 则直接赋值
					if i == allsub[lCount].MNum-1 {
						subchs = lastSub
					} else {
						//切分行数大于1时
						bsplit := regSplit.MatchString(allsub[lCount].SplitInfo[i].SSub)
						sEn := strings.Split(allsub[lCount].SplitInfo[i].SSub, ",")
						sChs := strings.Split(lastSub, "，")

						//有逗号结尾分隔符切分
						if (len(sChs) >= len(sEn)) && bsplit && preSplit {
							for j := range sEn {
								if j == len(sEn)-1 {
									break
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
							textten1 := []byte(allsub[lCount].SplitInfo[i].SSub)
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
										if (lpos <= avgline) && (j < len(CText) - 2){
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

								brnum := (len(allsub[lCount].SplitInfo) - i -1 ) * 2
								if k < ((len(CText) - brnum)) {
									subchs += CText[k]
								}  else  {
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
					allsub[lCount].SplitInfo[i].SCSub = subchs
					lastEnSub = lastEnSub[len(allsub[lCount].SplitInfo[i].SSub):len(lastEnSub)]
					lastSub = lastSub[len(subchs):len(lastSub)]
					subtext = strconv.Itoa(allsub[lCount].SplitInfo[i].SPos) + "\n" +
						allsub[lCount].SplitInfo[i].STime + "\n" +
						allsub[lCount].SplitInfo[i].SCSub + "\n" +
						allsub[lCount].SplitInfo[i].SSub + "\n"
					//fmt.Println(subtext)
					//fmt.Println("POS:" + strconv.Itoa(allsub[lCount].DPos))
					_, werr := subfile.WriteString(subtext)
					checkError(werr)
				}

			} else {
				if allsub[lCount].MNum == 1 {
					allsub[lCount].SplitInfo[0].SCSub = allsub[lCount].DCSub
					subtext = strconv.Itoa(allsub[lCount].SplitInfo[0].SPos) + "\n" +
						allsub[lCount].SplitInfo[0].STime + "\n" +
						allsub[lCount].DCSub + "\n" +
						allsub[lCount].SplitInfo[0].SSub + "\n"
					//fmt.Println(subtext)
					//fmt.Println("POS:" + strconv.Itoa(allsub[lCount].DPos))
					_, werr := subfile.WriteString(subtext)
					checkError(werr)

				}
			}
			lCount++
		}
		if slang == "en" {
			fmt.Println("A bilingual subtitle file has been generated ." + "\n")
			fmt.Println("Please check the file: " + trchsfilename + " ." + "\n")
		} else {
			fmt.Println("生成双语字幕文件." + "\n")
			fmt.Println("请查看文件: " + trchsfilename + " ." + "\n")
		}
		os.Chmod(trchsfilename, 0644)
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

	del_file(infilepath + ".json")

	jfile, _ := json.MarshalIndent(allsub, "", "\t")
	_ = ioutil.WriteFile(infilepath+".json", jfile, 0644)

}