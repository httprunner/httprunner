import setuptools
with open("README.md", "r") as fh:
    long_description = fh.read()
setuptools.setup(
    name="httprunner",  # 模块名称
    version="4.3.5",  # 当前版本
    author="diaodeng",  # 作者
    author_email="",  # 作者邮箱
    description="从原httprunner修改的",  # 模块简介
    long_description="从原httprunner修改用于支持httprunnermanager：https://gitee.com/zywstart/hrm2/tree/hrm4-newui/",  # 模块详细介绍
    long_description_content_type="text",  # 模块详细介绍格式
    url="https://github.com/diaodeng/httprunner/tree/master-diaodeng",  # 模块github地址
    packages=setuptools.find_packages(),  # 自动找到项目中导入的模块
    # 模块相关的元数据
    classifiers=[
        "Programming Language :: Python :: 3.7",
        "License :: OSI Approved :: MIT License",
    ],
    # 依赖模块
    install_requires=[
        "black >=22.3.0,<23.0.0",
        "Brotli >=1.0.9,<2.0.0",
        "Jinja2 >=3.0.3,<4.0.0",
        "jmespath >=0.9.5,<0.10.0",
        "loguru >=0.4.1,<0.5.0",
        "pydantic >=1.8,<1.9",
        "pytest >=7.1.1,<8.0.0",
        "pytest-html >=3.1.1,<4.0.0",
        "PyYAML >=6.0.1,<7.0.0",
        "requests >=2.31.0,<3.0.0",
        "sentry-sdk >=0.14.4,<0.15.0",
        "toml >=0.10.2,<0.11.0",
        "urllib3 >=1.26,<2.0",
    ],
    python_requires='>=3.7',
)