# uixt

From v4.3.0，HttpRunner will support mobile UI automation testing:

- iOS: based on [appium/WebDriverAgent], with forked client library [electricbubble/gwda] in golang
- Android: based on [appium-uiautomator2-server], with forked client library [electricbubble/guia2] in golang

Some UI recognition algorithms are also introduced for both iOS and Android:

- OpenCV: based on [OpenCV 4], with golang bindings [hybridgroup/gocv] and helper utils [electricbubble/gwda-ext-opencv]
- OCR: based on OCR API service from [volcengine], other API service may be extended

## Dependencies

### OpenCV

[OpenCV 4] should be pre-installed.

You can install OpenCV 4.6.0 using Homebrew on macOS.

```bash
$ brew install opencv
```

You can get more installation introduction on [hybridgroup/gocv].

### OCR

OCR API is a paid service, you need to pre-purchase and configure the environment variables.

- VEDEM_OCR_URL
- VEDEM_OCR_AK
- VEDEM_OCR_SK

## Thanks

This uixt module is initially forked from the following repos and made a lot of changes.

- [electricbubble/gwda-ext-opencv]
- [electricbubble/gwda]
- [electricbubble/guia2]


[electricbubble/gwda-ext-opencv]: https://github.com/electricbubble/gwda-ext-opencv
[appium/WebDriverAgent]: https://github.com/appium/WebDriverAgent
[electricbubble/gwda]: https://github.com/electricbubble/gwda
[electricbubble/guia2]: https://github.com/electricbubble/guia2
[OpenCV 4]: https://opencv.org/
[hybridgroup/gocv]: https://github.com/hybridgroup/gocv
[volcengine]: https://www.volcengine.com/product/text-recognition
[appium-uiautomator2-server]: https://github.com/appium/appium-uiautomator2-server
