---
type: docs
title: "Contribute to Symphony Docs"
linkTitle: "Contribute to Symphony Docs"
description: ""
weight: 100
---

Symphony documentation is built using [Hugo](https://gohugo.io/documentation/). The source code for the Symphony landing site is located in the `/landing` folder at the root of the repository, while the source code for the Symphony docs site is in the `/docs-site` folder. The Symphony GitHub build pipeline automatically builds and publishes both sites to:

- [https://eclipse-symphony.github.io/symphony-website/](https://eclipse-symphony.github.io/symphony-website/)
- [https://eclipse-symphony.github.io/docs/](https://eclipse-symphony.github.io/docs/)

## Prerequisites
* [Go](https://go.dev/doc/install) (1.22.3 or higher, latest stable version is recommended)
* [Hugo]((https://gohugo.io/documentation/)) (v0.148.1 or higher)

## Steps to Contribute to Documentation

Symphony aims to minimize and automate the contribution process so you can make meaningful improvements without the hassle of heavy procedures. To contribute, simply make your modifications and submit a pull request.

 1. **Create Your Branch and Clone Your Code**

    Create a separate branch for your changes and check it out to your local environment. If this is your first time setting up the project, run the following commands inside the `/docs-site` folder to ensure all Hugo modules are correctly loaded:

    ```bash
    hugo mod clean
    hugo mod tidy
    ```

 2. **Work on Docs Locally**

    To work on documentation locally, you’ll need to set up Go and Hugo. Then, navigate to the `/docs-site` folder and run:

    ```bash
    hugo server -D
    ```

    Then, you can open a browser and navigate to `http://localhost:1313/docs/` to preview the doc site. When you edit pages under the `/docs-site/content` folder, your can preview your changes live in the browser.

3. **Submitting a Pull Request**

    Once you're satisfied with your changes, create a pull request with a brief description of your modifications.    
