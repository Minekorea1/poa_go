#!/bin/bash
fyne bundle -package res -name NanumBarunGothicTTF fonts/NanumBarunGothic.ttf > bundle.go
fyne bundle -append      -name IconMain            image/icon_main.png >> bundle.go