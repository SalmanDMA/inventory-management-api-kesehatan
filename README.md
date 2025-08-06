<h1 align="center" style="font-size: 32px; font-weight: 800;"> Inventory APP API </h1> <br>

<p align='center' style="font-size: 16px; font-weight: 400;"> Reach Me On :</p>
 
<p align='center'>
  <a href="https://bit.ly/my-portofolio-salmandma">
    <img src="https://img.shields.io/badge/my_portfolio-000?style=for-the-badge&logo=ko-fi&logoColor=white" alt="portfolio">
  </a>
  <a href="https://www.linkedin.com/in/salmandma/">
    <img src="https://img.shields.io/badge/linkedin-0A66C2?style=for-the-badge&logo=linkedin&logoColor=white" alt="linkedin">
  </a>
  <a href="https://www.instagram.com/_slmndma_">
    <img src="https://img.shields.io/badge/instagram-E4405F?style=for-the-badge&logo=instagram&logoColor=white" alt="instagram">
  </a>
  <a href="https://github.com/SalmanDMA">
    <img src="https://img.shields.io/badge/github-181717?style=for-the-badge&logo=github&logoColor=white" alt="github">
  </a>
</p>

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

## Table of Contents

-   [Introduction](#introduction)
-   [Features](#features)
-   [Technologies](#technologies)
-   [Installation](#installation)
-   [Tests](#tests)
-   [Contributors](#contributors)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Introduction

[![On Process](https://img.shields.io/badge/build-on_process-blue)](https://github.com/SalmanDMA/alternatif-blog-api)
[![All Contributors](https://img.shields.io/badge/all_contributors-1-orange.svg?style=flat-square)](#contributors-)

Inventory App API is .... Go programming language using the Go Fiber framework.With Go Fiber, users can manage their accounts, authenticate with JWT, and manage roles.

## Features

-   User Management: Allows users to login, and manage their accounts.
-   Auth Login: Implements JWT authentication for secure login.
-   Role Management: Enables administrators to manage user roles and permissions.

## Technologies

-   Golang
-   Gorm
-   JWT
-   Go Fiber

## Installation

1. Clone Repository:

```bash
git clone https://github.com/SalmanDMA/go_fiber
```

2. Go to Directory:

```bash
cd go_fiber
```

3. Install Dependencies:

```bash
go mod tidy
```

4. Set Up Environment Variables:

Make sure you have set up your .env and .env.test files. Refer to .env.example for setting up these files.

5. Air Package

Make sure you have air package for live reloading

6. Init Air Package

```bash
air init
```

7.  Simple usage run Air Package

```bash
air -c .air.toml
```

8. Run Application with air (make sure you have PostgreSQL installed on your local machine):

```bash
air
```

9. Access Applications:

Use an API testing tool like Postman to access the project APIs.

## Tests

1. Ensure you have set up your .env.test file. Refer to .env.example for details.

2. Run Tests with Coverage:

```bash
go test -cover ./...
```

3. Generate Coverage Report:

```bash
go test -v ./... -coverprofile=coverage.out
```

4. Generate HTML Coverage Report:

```bash
go tool cover -html=coverage.out -o coverage.html
```

## Contributors

This project is maintained by [Salman DMA](https://github.com/SALMANDMA).
