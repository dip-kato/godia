# godia

**AWSの超シンプルなネットワーク構成図作ってくれるやつ**。EC2と以下の関係性に全振りしてます。他の情報は載りません。

- **サブネットの種類**
- **ゲートウェイやサブネットをまたいだ通信の向き**
- **セキュリティグループ**

# インストール方法

パスの通ったところに置く場合

```
go get github.com/dip-kato/godia
```

クローンしてビルドする場合

```
git clone https://github.com/dip-kato/godia
cd godia
go build .
```

[バイナリ落として即使いたいぜ！の人はこっち](https://github.com/dip-kato/godia/releases)

# 使い方

## ネットワーク情報を引っこ抜く

**awscliがセットアップされていてインスタンスの情報にアクセス権があるcliがあるところからネットワーク情報をシェルでひっこ抜きます**

![image](https://user-images.githubusercontent.com/95202883/143992948-d95b3f33-6ead-463f-b426-86cd002d7324.png)
<br>
伏字が多くてわかり辛いですが、EC2についている各SGが出力されていると思ってください

```
getNW.sh> ([1] profile) ([2] Tag Name) ([3] Tag Value)
```

こういうオプション指定です。**[2][3]はNameタグが[3]だったら、、みたいに指定します**。つまりName以外の別のタグを指定して収集できます

## 出力結果をファイル化する

上記シェルの出力をファイルにリダイレクト、または手動でファイルに落とします。<br>

![image](https://user-images.githubusercontent.com/95202883/143993303-c99217fb-8085-48fd-a8f9-c8090d1508df.png)

ここで注意するのは**ネットワーク階層が上位なものほどファイル名を上位にします**[0-9,a-z,A-Z順ってやつ]。一つのフォルダ[datがデフォ]に格納します。<br>
この例は1_webが一階層目、2_apiが二階層目、、のようになっています。<br>

## 実行します

go-diagramsというフォルダに各種アセット(※アイコンのこと)と、**DOT言語が書かれてファイルが生成されます**<br>

## 図化します

[Graphvizで図化します(https://graphviz.org/)

```
dot.exe -Tpng .\go-diagrams\nw.dot > nw.png
```

こんなかんじでネットワーク構成図が出力されます

![image](https://user-images.githubusercontent.com/95202883/143993526-91f5bf28-bd91-46f2-bb09-bb8e67efd280.png)

# 質疑応答
## 3階層のみしか対応しないんですか？

![image](https://user-images.githubusercontent.com/95202883/143993593-30ba60d8-83a6-4552-9439-8f84463057cf.png)

階層ファイルを増やしていけば**何層でも**イケますよ！

## Internet Gatewayとか外部との通信ルートも出したい

**.iniファイル(デフォはgodia.ini)に外部と通信するSGを定義して紐づけてあげれば描けます！**

![image](https://user-images.githubusercontent.com/95202883/143993727-81264698-fdfa-492e-9aaf-322dd45d70b1.png)

```
1_web,I,sg_xxxx,GW1
```

### .iniファイルの書き方

(紐づけたいファイル),(通信タイプ),(紐づけたいSG名),(Internet Gatewayの名前、ID)<br>
通信タイプは I: inbound>内部向け通信 O: outbound>外部向け通信 D: double>双方向 で定義します<br>

# オプション

```
Usage of godia.exe:
  -debug
        [-debug=debug mode (true is enable)]
  -dir string
        [-dir=search directory] (default "dat")
  -ini string
        [-ini=configuration file (.ini) filename] (default "godia.ini")
  -output string
        [-output=output .dot filename] (default "nw")
  -verbise
        [-verbose=incude id verbose (true is enable)] (default true)
  -vpc, string
        [-vpc=vpc name and id (for Label)] (default "vpc,vpc-00000000000000000")
```

## -debug

デバッグモードです。オンにすると色々出力されます

## -dir
インスタンスのネットワーク情報ファイルを格納するディレクトリ名を指定できます

## -ini

外部ネットワークとの通信紐づけコンフィグ名を指定できます

## -output

アウトプットするDOT言語のファイル名を指定できます

## -verbise

Subnet名だけじゃなくて、SubnetIDも一緒に出力するモードです

## -vpc

VPC名を指定できます

# ライセンス
MIT License
