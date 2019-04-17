# SubtitleTranslation
提高机翻中文字幕准确性

# 机翻中文字幕辅助软件 / 版本 : TrSubtitle / 0.1 
####  by jikai   Email:jikaimail@gmail.com
### 本软件调用：
#### 【https://github.com/huichen/sego  sego Go中文分词】进行中文字幕的分割
#### 【https://github.com/gitote/chardet  chardet】判断输入的文件字符集

## 使用方法：
###     很多视频由于某种原因，无人工翻译的中文字幕；本软件通过对原文字幕进行处理，提高机翻中文字幕的准确性。 也可作为人工中文字幕翻译简单的辅助工具。
### 1) TrSubtitle -infile 字幕文件名  
###    生成要翻译的英文文件；
### 2) 将生成的待翻译的英文文件，人工翻译并存储在一个文件内；
###    确保翻译内容与原内容的行位置和总行数要匹配。
### 3) TrSubtitle -infile 字幕文件名  -trfile 已翻译的文件名  
###    生成最终双语字幕文件。
### 4) TrSubtitle -jsfile json文件名
###    如果需要对字幕进一步调整，可在json文件内对字幕内容进行修正；
###    程序根据调整后的json文件，重新生成双语字幕文件。   
### 注意事项：确保翻译内容与原内容的行位置和总行数要匹配。 

## 参数选项:
###  -h          : 帮助
###  -lang       : chs显示中文帮助 en显示英文帮助. 默认chs.
###  -infile     : 输入要处理的字幕文件名. 
###               (需要无格式的srt字幕文件)
###  -trfile     : 输入已翻译文件名. 
###  -jsfile     : 输入json文件名.   

## 推荐的机翻网址：
### https://translate.google.com/
### https://cn.bing.com/Translator
### https://fanyi.baidu.com/

## 机翻中文字幕有用的工具链接： 
###  1) 对于无英文字幕的视频可采用【https://github.com/agermanidis/autosub Autosub】  生成英文字幕
###  2) 推荐一个字幕编辑器 【https://github.com/SubtitleEdit/subtitleedit  字幕编辑器】
##     也许大家注意到了，这两个软件均具有中文字幕翻译功能；但中文翻译的效果并不理想；
##  这就是TrSubtitle软件编写的目的和想改进的内容，具体的改进的效果则限于个人水平啦！
## 衷心希望此软件能对你有帮助。
