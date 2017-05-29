import React, { Component } from 'react';

class ResultList extends Component {
    render() {
        var key = 0;
        if (this.props.files.length === 0) {
            return (
                <div>
                    <div className="result-count-info">Brak wyników</div>
                </div>
            );
        }
        var trs = [];
        trs.push(<div key={key++} className="result-count-info">Znaleziono następującą ilość pasujących wyników: <span style={{'fontWeight': 'bold'}}>{this.props.files.length}</span></div>);
        var textToDiv = (text) => {
            return <div key={key++}>{"..." + text + "..."}</div>;
        };

        // sort fix, assuming latest version of backend
        function compare(a,b) {
            if (a.Score.S < b.Score.S) {
                return 1;
            }
            if (a.Score.S > b.Score.S) {
                return -1;
            }
            return 0;
        }
        // new field/property 'Score'
        if (this.props.files.hasOwnProperty('Score')) {
          this.props.files.sort(compare);
        }

        for (var file of this.props.files) {
            var title = file.Filename;
            var titleRest = '';
            var indexOfId = file.Filename.indexOf("z dnia");
            if (indexOfId !== -1) {
                title = file.Filename.slice(0, indexOfId);
                titleRest = file.Filename.slice(indexOfId);
            }
            else {
                // Old implementation, may be subject to change. Used as a last chance resort here
                var i = 0;
                var titleEnded = false;
                var fileNameLength = Object.keys(file.Filename).length;
                while (!titleEnded) {
                    var character = file.Filename[i];
                    var nextCharacter = null;
                    if (i + 1 < fileNameLength) {
                        nextCharacter = file.Filename[i + 1];
                    }
                    if (nextCharacter === null) {
                        // it is the last character in string
                        i++;
                        titleEnded = true;
                    }
                    else {
                        // it is not the last character
                        if (character === ' ' || (character === character.toUpperCase() && (nextCharacter === ' ' || nextCharacter === nextCharacter.toUpperCase()))) {
                            i++
                        }
                        else {
                            titleEnded = true
                        }
                    }
                }
                if (i) {
                    title = file.Filename.slice(0, i);
                    titleRest = file.Filename.slice(i);
                }
            }

            var fileDate = new Date(file.Date);
            var day = ("0" + fileDate.getUTCDate()).slice(-2);
            var month = ("0" + (fileDate.getUTCMonth() + 1)).slice(-2); //months from 1-12
            var year = fileDate.getUTCFullYear();

            trs.push(<div key={key++} className="font-weight-bold">{title}</div>);
            trs.push(<div key={key++}>{titleRest}</div>);
            trs.push(<div key={key++}><a className="red-link" target="_blank" href={file.Link}>{file.Link}</a><span> </span><a className="blue-link" href={"/api/download?id=" + file.ID} download>[zobacz kopię lokalną]</a></div>);
            trs.push(<div key={key++}>{file.Texts.map(textToDiv)}</div>);
            trs.push(<div key={key++} className="text-smaller">| rok wydania: {file.Year} | data wydania: {year + "-" + month + "-" + day} | źródło: {file.Koala} | pozycja: {file.Position} |</div>);
            trs.push(<div key={key++} className="search-title"></div>);
        }
        return (
            <div id="resultList">
                {trs}
            </div>
        );
    }
}

export default ResultList;
