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
        "Operating System :: OS Independent",
    ],
    # 依赖模块
    install_requires=[
        'pillow',
    ],
    python_requires='>=3.7',
)