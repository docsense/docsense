import React, { Component } from 'react';
import './App.css';
import Header from './Header/Header'
import Search from './Search.jsx';
import ManageSpLists from './ManageSpLists.jsx';

class App extends Component {
    render() {
        var component;
        if (window.location.hash === "#manage_sp_lists") {
            component = <ManageSpLists />;
        } else {
            component = <Search />;
        }
        return (
            <div>
                <div className="container my-container">
                    <Header />
                    {component}
                </div>
            </div>
        );
    }
}

export default App;
