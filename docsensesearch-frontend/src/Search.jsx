import React, { Component } from 'react';
import {OverlayTrigger, Tooltip} from 'react-bootstrap';
import ResultList from './ResultList';
import LogoImg from './Images/logo.svg';
import Footer from './Footer/Footer';

class UploadFile extends Component {
    constructor(props) {
        super(props);
        this.handleChange = this.handleChange.bind(this);
        this.handleClick = this.handleClick.bind(this);
    }
    handleClick() {
        var input = document.getElementById("fileinput");
        input.click();
    }
    handleChange() {
        var input = document.getElementById("fileinput");
        var form = document.getElementById('fileform');

        var formData = new FormData(form);
        var xhr = new window.XMLHttpRequest();
        var that = this;
        that.props.setLoading(true);
        xhr.addEventListener("load", function(event) {
            var files = JSON.parse(event.currentTarget.response);
            //console.log(files);
            for (var i = 0; i < files.length; i++) {
                files[i].Texts = [];
            }
            that.props.setFiles(files);
        });
        xhr.open("POST", "/api/searchbyfile");
        xhr.send(formData);
        input.value = "";

    }
    render() {
        const tooltipFile = (
            <Tooltip id="tooltipFile">Wyszukaj plikiem</Tooltip>
        );

        return (
            <div>
                <OverlayTrigger placement="top" overlay={tooltipFile}>
                    <i className="fa fa-lg form-control-feedback  icon-wyszukiwanie_plikiem" aria-hidden="true" onClick={this.handleClick}></i>
                </OverlayTrigger>
                <form id="fileform" style={{display: "none"}}>
                    <input id="fileinput" type="file" accept="application/pdf,application/msword" name="file" onChange={this.handleChange} />
                </form>
            </div>
        )
    }
}

class Search extends Component {
    constructor(props) {
        super(props);
        this.state = {value: '', state: "idle", files: [], sort: ''};

        this.searched = false;
        this.handleChange = this.handleChange.bind(this);
        this.handleSearchClick = this.handleSearchClick.bind(this);
        this.setFiles = this.setFiles.bind(this);
        this.setLoading = this.setLoading.bind(this);
        //this.doSort = this.doSort.bind(this);
    }
    handleChange(event) {
        this.setState({value: event.target.value}, function () {
            if (this.state.value && this.state.state !== 'loading') {
                clearTimeout(this.timeout);
                this.timeout = setTimeout(() => { this.doSearch(); }, 1000);
            }
        });
    }
    handleSearchClick(event) {
        if (this.state.value && this.state.state !== 'loading') {
            clearTimeout(this.timeout);
            this.doSearch();
        }
    }
    setFiles(files) {
        this.searched = true;
        this.setState({files: files});
        //this.setState({state: "idle"});
        this.setLoading(false);
    }
    setLoading(value) {
        if (value) {
            this.setState({state: "loading"});
        }
        else {
            this.setState({state: "idle"});
        }
    }
    doSearch() {
        //this.setState({state: 'loading'});
        this.setLoading(true);
        var form = new FormData();
        form.append("usos_id", this.props.usos_id);
        form.append("token", this.props.token);
        form.append("phrase", this.state.value);
        form.append("searched_folder", this.props.searchedFolder);
        fetch('/api/search', {
            method: 'POST',
            body: form}
            ).then((data) => data.json()).then((data) => {
                console.log(data);
                this.setFiles(data);
            }
        );
    }
    doSort(propertyId) {
        switch(propertyId) {
            case 0:
                if (this.state.sort === '0+') {
                    this.setState({files: this.state.files.sort(this.dynamicSort("-Year"))});
                    this.setState({sort: '0-'});
                }
                else {
                    this.setState({files: this.state.files.sort(this.dynamicSort("Year"))});
                    this.setState({sort: '0+'});
                }
                break;
            case 1:
                if (this.state.sort === '1+') {
                    this.setState({files: this.state.files.sort(this.dynamicSort("-Date"))});
                    this.setState({sort: '1-'});
                }
                else {
                    this.setState({files: this.state.files.sort(this.dynamicSort("Date"))});
                    this.setState({sort: '1+'});
                }
                break;
            case 2:
                if (this.state.sort === '2+') {
                    this.setState({files: this.state.files.sort(this.dynamicSort("-Koala"))});
                    this.setState({sort: '2-'});
                }
                else {
                    this.setState({files: this.state.files.sort(this.dynamicSort("Koala"))});
                    this.setState({sort: '2+'});
                }
                break;
            case 3:
                if (this.state.sort === '3+') {
                    this.setState({files: this.state.files.sort(this.dynamicSort("-Position"))});
                    this.setState({sort: '3-'});
                }
                else {
                    this.setState({files: this.state.files.sort(this.dynamicSort("Position"))});
                    this.setState({sort: '3+'});
                }
                break;
            default:
                this.setState({sort: ''});
                console.log('Error in sorting!');
        }
        console.log(this.state.files);
    }
    dynamicSort(property) {
        var sortOrder = 1;
        if(property[0] === "-") {
            sortOrder = -1;
            property = property.substr(1);
        }
        return function (a,b) {
            var result = (a[property] < b[property]) ? -1 : (a[property] > b[property]) ? 1 : 0;
            return result * sortOrder;
        }
    }
    renderPage(filesCount) {
        const tooltipSearch = (
            <Tooltip id="tooltipSearch">Wyszukaj słowem</Tooltip>
        );

        if (!filesCount && !this.searched) {
            return (
                <div>
                    <div className="row">
                        <div className="col-sm-12 text-align-center">
                            <img id="appBodyImg" src={LogoImg} alt="mainLogo"/>
                        </div>
                    </div>
                    <div className="row">
                        <div className="col-sm-6 col-centered">
                            <div className="form-group has-feedback">
                                <input type="text" className="form-control" value={this.state.value} onChange={this.handleChange} placeholder='Szukaj słowem lub plikiem...' autoFocus/>
                                <OverlayTrigger placement="top" overlay={tooltipSearch}>
                                    <i className="fa fa-lg form-control-feedback icon-wyszukiwanie_slowem" aria-hidden="true" style={{right: '34px'}} onClick={this.handleSearchClick}></i>
                                </OverlayTrigger>
                                <UploadFile setFiles={this.setFiles} setLoading={this.setLoading}/>
                            </div>
                        </div>
                    </div>
                </div>
            );
        }
        else {
            return (
                <div>
                    <div className="row">
                        <div className="col-sm-12">
                            <div className="form-group has-feedback">
                                <input type="text" className="form-control" value={this.state.value} onChange={this.handleChange} placeholder='Szukaj słowem lub plikiem...' autoFocus/>
                                <OverlayTrigger placement="top" overlay={tooltipSearch}>
                                    <i className="fa fa-lg form-control-feedback icon-wyszukiwanie_slowem" aria-hidden="true" style={{right: '34px'}} onClick={this.handleSearchClick}></i>
                                </OverlayTrigger>
                                <UploadFile setFiles={this.setFiles} setLoading={this.setLoading} />
                            </div>
                        </div>
                    </div>
                    <div className="row">
                        <div className="col-sm-12">
                            <span>Filtruj po: </span>
                            <a className="blue-link clickable" onClick={this.doSort.bind(this, 0)}>
                                roku wydania
                                {this.state.sort==='0+' ? <i className="fa fa-arrow-down" aria-hidden="true"></i> : ''}
                                {this.state.sort==='0-' ? <i className="fa fa-arrow-up" aria-hidden="true"></i> : ''}
                            </a>|
                            <a className="blue-link clickable" onClick={this.doSort.bind(this, 1)}>
                                dacie wydania
                                {this.state.sort==='1+' ? <i className="fa fa-arrow-down" aria-hidden="true"></i> : ''}
                                {this.state.sort==='1-' ? <i className="fa fa-arrow-up" aria-hidden="true"></i> : ''}
                            </a>|
                            <a className="blue-link clickable" onClick={this.doSort.bind(this, 2)}>
                                źródle
                                {this.state.sort==='2+' ? <i className="fa fa-arrow-down" aria-hidden="true"></i> : ''}
                                {this.state.sort==='2-' ? <i className="fa fa-arrow-up" aria-hidden="true"></i> : ''}
                            </a>|
                            <a className="blue-link clickable" onClick={this.doSort.bind(this, 3)}>
                                pozycji
                                {this.state.sort==='3+' ? <i className="fa fa-arrow-down" aria-hidden="true"></i> : ''}
                                {this.state.sort==='3-' ? <i className="fa fa-arrow-up" aria-hidden="true"></i> : ''}
                            </a>
                        </div>
                        <div className="col-sm-12">
                            <ResultList files={this.state.files} showFile={this.props.showFile} />
                        </div>
                    </div>
                </div>
            );
        }
    }
    render() {
        var results = this.renderPage(this.state.files.length);
        return (
            <div>
                {results}
                <Footer loading={this.state.state}/>
            </div>
        );
    }
}

export default Search;
