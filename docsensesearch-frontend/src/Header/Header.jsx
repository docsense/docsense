import React, { Component } from 'react';
import './Header.css';
import HeaderLogoImg from '../Images/logo2.svg';

class Header extends Component {
    constructor(props){
        super(props);
        this.refresh = this.refresh.bind(this);
    }

    refresh(e){   // function upvote
        e.preventDefault();
        window.location.reload();
        return false

    }
    render() {
        return (
            <div id="header" className="row">
                <div className="col-xs-12">
                    <img id="headerLogoImg" className="clickable" src={HeaderLogoImg} alt="Header Logo" onClick={this.refresh.bind(this)}/>
                </div>
            </div>
        );
    }
}

export default Header;
