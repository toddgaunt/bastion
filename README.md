# Bastion
Bastion is a simple webserver for serving articles written in markdown. These markdown
articles are converted to html automatically as they are updated, and have a special
header section to specify useful metadata about the article to Bastion.

## Website layout
```
www.example.com/
    config.json
    content/
        about.md
        contact.md
    static/
        default.css
    templates/
        article.html
        index.html
        problem.html
```

## Developer Quickstart
First, make sure the following programs are installed:
- go
- pigz
- pv
- realpath
- tar

Then run the following commands:
```
./make.sh build
./bastion www.example.com
```
For commandline options and usage information run `bastion -h`
