# -*- coding: utf-8 -*-
from setuptools import setup

packages = \
['httprunner',
 'httprunner.app',
 'httprunner.app.routers',
 'httprunner.builtin',
 'httprunner.ext',
 'httprunner.ext.har2case',
 'httprunner.ext.locust',
 'httprunner.ext.uploader']

package_data = \
{'': ['*']}

install_requires = \
['black>=19.10b0,<20.0',
 'jinja2>=2.10.3,<3.0.0',
 'jmespath>=0.9.5,<0.10.0',
 'loguru>=0.4.1,<0.5.0',
 'pydantic>=1.4,<2.0',
 'pytest-html>=2.1.1,<3.0.0',
 'pytest>=5.4.2,<6.0.0',
 'pyyaml>=5.1.2,<6.0.0',
 'requests>=2.22.0,<3.0.0',
 'sentry-sdk>=0.14.4,<0.15.0']

extras_require = \
{'allure': ['allure-pytest>=2.8.16,<3.0.0'],
 'locust': ['locust>=1.0.3,<2.0.0'],
 'upload': ['requests-toolbelt>=0.9.1,<0.10.0', 'filetype>=1.0.7,<2.0.0']}

entry_points = \
{'console_scripts': ['har2case = httprunner.cli:main_har2case_alias',
                     'hmake = httprunner.cli:main_make_alias',
                     'hrun = httprunner.cli:main_hrun_alias',
                     'httprunner = httprunner.cli:main',
                     'locusts = httprunner.ext.locust:main_locusts']}

setup_kwargs = {
    'name': 'httprunner',
    'version': '3.1.4',
    'description': 'One-stop solution for HTTP(S) testing.',
    'long_description': '\n# HttpRunner\n\n[![downloads](https://pepy.tech/badge/httprunner)](https://pepy.tech/project/httprunner)\n[![unittest](https://github.com/httprunner/httprunner/workflows/unittest/badge.svg\n)](https://github.com/httprunner/httprunner/actions)\n[![integration-test](https://github.com/httprunner/httprunner/workflows/integration_test/badge.svg\n)](https://github.com/httprunner/httprunner/actions)\n[![codecov](https://codecov.io/gh/httprunner/httprunner/branch/master/graph/badge.svg)](https://codecov.io/gh/httprunner/httprunner)\n[![pypi version](https://img.shields.io/pypi/v/httprunner.svg)](https://pypi.python.org/pypi/httprunner)\n[![pyversions](https://img.shields.io/pypi/pyversions/httprunner.svg)](https://pypi.python.org/pypi/httprunner)\n[![TesterHome](https://img.shields.io/badge/TTF-TesterHome-2955C5.svg)](https://testerhome.com/github_statistics)\n\n*HttpRunner* is a simple & elegant, yet powerful HTTP(S) testing framework. Enjoy! ✨ 🚀 ✨\n\n## Design Philosophy\n\n- Convention over configuration\n- ROI matters\n- Embrace open source, leverage [`requests`][requests], [`pytest`][pytest], [`pydantic`][pydantic], [`allure`][allure] and [`locust`][locust].\n\n## Key Features\n\n- Inherit all powerful features of [`requests`][requests], just have fun to handle HTTP(S) in human way.\n- Define testcase in YAML or JSON format, run with [`pytest`][pytest] in concise and elegant manner. \n- Record and generate testcases with [`HAR`][HAR] support.\n- Supports `variables`/`extract`/`validate`/`hooks` mechanisms to create extremely complex test scenarios.\n- With `debugtalk.py` plugin, any function can be used in any part of your testcase.\n- With [`jmespath`][jmespath], extract and validate json response has never been easier.\n- With [`pytest`][pytest], hundreds of plugins are readily available. \n- With [`allure`][allure], test report can be pretty nice and powerful.\n- With reuse of [`locust`][locust], you can run performance test without extra work.\n- CLI command supported, perfect combination with `CI/CD`.\n\n## Sponsors\n\nThank you to all our sponsors! ✨🍰✨ ([become a sponsor](docs/sponsors.md))\n\n### 金牌赞助商（Gold Sponsor）\n\n[<img src="docs/assets/hogwarts.png" alt="霍格沃兹测试学院" width="400">](https://ceshiren.com/)\n\n> [霍格沃兹测试学院](https://ceshiren.com/) 是业界领先的测试开发技术高端教育品牌，隶属于测吧（北京）科技有限公司。学院课程均由 BAT 一线测试大咖执教，提供实战驱动的接口自动化测试、移动自动化测试、性能测试、持续集成与 DevOps 等技术培训，以及测试开发优秀人才内推服务。[点击学习!](https://ke.qq.com/course/254956?flowToken=1014690)\n\n霍格沃兹测试学院是 HttpRunner 的首家金牌赞助商。\n\n### 开源服务赞助商（Open Source Sponsor）\n\n[<img src="docs/assets/sentry-logo-black.svg" alt="Sentry" width="150">](https://sentry.io/_/open-source/)\n\nHttpRunner is in Sentry Sponsored plan.\n\n## Subscribe\n\n关注 HttpRunner 的微信公众号，第一时间获得最新资讯。\n\n![](docs/assets/qrcode.jpg)\n\n[requests]: http://docs.python-requests.org/en/master/\n[pytest]: https://docs.pytest.org/\n[pydantic]: https://pydantic-docs.helpmanual.io/\n[locust]: http://locust.io/\n[jmespath]: https://jmespath.org/\n[allure]: https://docs.qameta.io/allure/\n[HAR]: http://httparchive.org/\n\n\n',
    'author': 'debugtalk',
    'author_email': 'debugtalk@gmail.com',
    'maintainer': None,
    'maintainer_email': None,
    'url': 'https://github.com/httprunner/httprunner',
    'packages': packages,
    'package_data': package_data,
    'install_requires': install_requires,
    'extras_require': extras_require,
    'entry_points': entry_points,
    'python_requires': '>=3.6,<4.0',
}


setup(**setup_kwargs)
