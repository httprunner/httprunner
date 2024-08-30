# uixt

From v4.3.0，HttpRunner will support mobile UI automation testing:

- iOS: based on [appium/WebDriverAgent], with forked client library [electricbubble/gwda] in golang
- Android: based on [appium-uiautomator2-server], with forked client library [electricbubble/guia2] in golang

Some UI recognition algorithms are also introduced for both iOS and Android:

- OCR: based on OCR API service from [volcengine], other API service may be extended

## Dependencies

### OCR

OCR API is a paid service, you need to pre-purchase and configure the environment variables.

- VEDEM_IMAGE_URL
- VEDEM_IMAGE_AK
- VEDEM_IMAGE_SK

## Thanks

This uixt module is initially forked from the following repos and made a lot of changes.

- [electricbubble/gwda]
- [electricbubble/guia2]


[appium/WebDriverAgent]: https://github.com/appium/WebDriverAgent
[electricbubble/gwda]: https://github.com/electricbubble/gwda
[electricbubble/guia2]: https://github.com/electricbubble/guia2
[volcengine]: https://www.volcengine.com/product/text-recognition
[appium-uiautomator2-server]: https://github.com/appium/appium-uiautomator2-server
