import React, { Component } from 'react';
import './Footer.css';
import FooterLoadingImg from '../Images/gif.gif';
import FooterLogoImg from '../Images/logo3.svg'


class Footer extends Component {
    render() {
        var source = FooterLogoImg;
        if (this.props.loading === 'loading') {
            source = FooterLoadingImg;
        }
        else {
            source = FooterLogoImg;
        }
        return (
            <div id="footer" className="navbar navbar-default navbar-fixed-bottom footer-message">
                <img id="footerLogoImg" src={source} alt="footerLogo"/>
                <span id="footerCaption"><a href="http://www.uw.edu.pl" className="footer-link clickable">Uniwersytet Warszawski</a> |  <a href="//dziennik.uw.edu.pl" className="footer-link clickable">Dziennik UW</a> | <a href="//monitor.uw.edu.pl" className="footer-link clickable">Monitor UW</a></span>
            </div>
        );
    }
}

export default Footer;
