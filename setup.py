#encoding: utf-8
import io

from httprunner import __version__
from setuptools import find_packages, setup

with io.open("README.rst", encoding='utf-8') as f:
    long_description = f.read()

install_requires = open("requirements.txt").readlines()

setup(
    name='HttpRunner',
    version=__version__,
    description='HTTP test runner, not just about api test and load test.',
    long_description=long_description,
    author='Leo Lee',
    author_email='mail@debugtalk.com',
    url='https://github.com/debugtalk/HttpRunner',
    license='MIT',
    packages=find_packages(exclude=['test.*', 'test']),
    package_data={
        'httprunner': ['locustfile_template'],
    },
    keywords='api test',
    install_requires=install_requires,
    classifiers=[
        "Development Status :: 3 - Alpha",
        'Programming Language :: Python :: 2.7',
        'Programming Language :: Python :: 3.4',
        'Programming Language :: Python :: 3.5',
        'Programming Language :: Python :: 3.6'
    ],
    entry_points={
        'console_scripts': [
            'ate=httprunner.cli:main_ate',
            'httprunner=httprunner.cli:main_ate',
            'hrun=httprunner.cli:main_ate',
            'locusts=httprunner.cli:main_locust'
        ]
    }
)
