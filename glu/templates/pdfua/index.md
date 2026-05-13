---
title: Accessible Document
author: Your Name
lang: en
format: PDF/UA
papersize: a4
---

# Accessibility Demo

This document targets PDF/UA-1 compliance. The frontmatter enables
the tagged-PDF pipeline (`format: PDF/UA`); the `lang:` field sets
the document language for screen readers.

## Structure

Headings (h1–h6) produce a structure tree consumable by assistive
technology. Use them hierarchically — don't skip levels.

## Lists

Bulleted and numbered lists become L/LI structure elements:

- First item
- Second item
- Third item

## Tables

Pipe tables get TH/TD tags with header row detection:

| Region | Population |
| ------ | ---------: |
| North  |  1,250,000 |
| South  |    980,000 |
| East   |    640,000 |

## Images

Place images with descriptive alt text:

`![A bar chart showing 2024 sales by region](sales.png)`

The alt attribute becomes the `/Alt` entry on the Figure structure
element. Decorative images (no `alt=`) are marked as artifacts.

## Verification

Validate output with veraPDF:

    verapdf --flavour ua1 index.pdf
