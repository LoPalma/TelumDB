from setuptools import setup, find_packages, Extension
from pybind11.setup_helpers import Pybind11Extension, build_ext
from pybind11 import get_cmake_dir
import pybind11

# Define the extension module
ext_modules = [
    Pybind11Extension(
        "telumdb._core",
        [
            "src/core.cpp",
            # Add more C++ source files here
        ],
        include_dirs=[
            pybind11.get_include(),
            "../../internal",
            "../../pkg",
        ],
        cxx_std=17,
        define_macros=[("VERSION_INFO", '"dev"')],
    ),
]

setup(
    name="telumdb",
    version="0.1.0",
    author="TelumDB Contributors",
    author_email="contributors@telumdb.io",
    description="Python client for TelumDB - The World's First Hybrid General-Purpose + AI Tensor Database",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    url="https://github.com/telumdb/telumdb",
    ext_modules=ext_modules,
    cmdclass={"build_ext": build_ext},
    packages=find_packages(where="src"),
    package_dir={"": "src"},
    python_requires=">=3.8",
    install_requires=[
        "numpy>=1.20.0",
        "pandas>=1.3.0",
        "psycopg2-binary>=2.9.0",
        "requests>=2.25.0",
        "pydantic>=1.8.0",
    ],
    extras_require={
        "dev": [
            "pytest>=6.0",
            "pytest-cov>=2.0",
            "black>=21.0",
            "flake8>=3.9",
            "mypy>=0.910",
            "sphinx>=4.0",
            "sphinx-rtd-theme>=0.5",
        ],
        "ml": [
            "scikit-learn>=1.0",
            "torch>=1.9",
            "tensorflow>=2.6",
        ],
    },
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "Intended Audience :: Science/Research",
        "License :: OSI Approved :: Apache Software License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: C++",
        "Topic :: Database",
        "Topic :: Scientific/Engineering :: Artificial Intelligence",
    ],
    keywords="database tensor ai machine-learning sql vector",
    project_urls={
        "Bug Reports": "https://github.com/telumdb/telumdb/issues",
        "Source": "https://github.com/telumdb/telumdb",
        "Documentation": "https://telumdb.readthedocs.io/",
    },
)