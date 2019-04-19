# 机翻双语字幕辅助软件/版本：TrSubtitle / 0.3 
####  by jikai   Email:jikaimail@gmail.com
# 软件目的：
## 通过分句将原文字幕全部提取为待译文件，完整的待译文件可提高机翻的准确性；采用分词、分句相结合的方式对译文按原时间轴进行分割，提高译文字幕的观看舒适度。
## 此软件不是翻译软件、AI翻译 也不进行任何翻译工作，它是一个提高机翻双语字幕准确性的辅助工具。
## 使用方法：
###     很多视频由于某种原因，无所需译文字幕；此软件用于将无格式的srt原文字幕中的原文内容提取为待译原文；使用者通过机翻网站或者人工翻译将待译原文翻译成为译文；软件再将译文与原文字幕进行分析处理并合并生成双语字幕。
### 1) TrSubtitle -infile 原文字幕文件名  
###    生成字幕原文待译文件；
### 2) 将生成的待译原文文件通过机翻网站或者人工翻译后存储在一个文件内；
###    确保译文与原文内容的行位置和总行数要匹配。
### 3) TrSubtitle -infile 原文字幕文件名  -trfile 译文文件名  
###    生成所需的字幕文件。
### 4) TrSubtitle -jsfile json文件名
###    如果需要对字幕进一步调整，可在json文件内对字幕内容进行修正；
###    程序根据调整后的json文件，重新生成所需的字幕文件。   

## 参数选项:
###  -h          : 帮助
###  -lang       : chs显示中文帮助 en显示英文帮助. 默认chs.
###  -infile     : 输入要处理的原文字幕文件名.  (需要无格式的srt字幕文件)
###  -trfile     : 输入译文文件名. 
###  -jsfile     : 输入json文件名.
###  -stype      : o 仅生成译文字幕 b 生成双语字幕文件 默认b.  

## 推荐的机翻网址：
### https://translate.google.com/
### https://cn.bing.com/Translator
### https://fanyi.baidu.com/

### 本软件调用：
#### 【https://github.com/huichen/sego  sego Go中文分词】进行中文字幕的分割
#### 【https://github.com/gitote/chardet  chardet】判断输入的文件字符集
###  本软件完全使用golang 1.11+ 开发

## 其它机翻中文字幕工具链接： 
###  1) 对于无英文字幕的视频可采用【https://github.com/agermanidis/autosub Autosub】  此软件通过谷歌的语音识别引擎可生成英文srt字幕文件
###  2) 推荐一个字幕编辑器 【https://github.com/SubtitleEdit/subtitleedit  字幕编辑器】
##     也许大家注意到了，这两个软件均具有中文字幕翻译功能；但中文翻译的效果并不理想；
##  这就是TrSubtitle软件编写的目的和想改进的内容，具体的改进的效果则限于个人水平啦！
## 衷心希望此软件能对你有帮助。
