# autoTyper

支持通过热键来快速输入文本，还支持通过`f10`打开特殊输入来支持像vnc远程服务器时自动输入密码的类似操作。

![image-20221221152108154](https://raw.githubusercontent.com/GavinTan/files/master/picgo/image-20221221152108154.png)

## build

~~~shell
#fyne-cross需要docker
go install github.com/gavintan/fyne-cross@latest

fyne-cross windows
fyne-cross darwin -app-id com.tw.autoTyper  -app-version 1.0.2
~~~

