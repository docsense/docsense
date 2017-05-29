import React, { Component } from 'react';
import { Form, Table } from 'semantic-ui-react';

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

class ManageSpLists extends Component {
    constructor(props) {
        super(props);
        this.submit = this.submit.bind(this);
        this.login = this.login.bind(this);
        this.state = {lists: [], message: ""};
    }
    login() {
        var login = document.getElementById("login").firstChild.value;
        var password = document.getElementById("password").firstChild.value;
        this.setState({login: login, password: password});
        sleep(100).then(()=>{
            this.updateSPLists();
        });
    }
    updateSPLists() {
        var that = this;
        fetch('/api/managesplists', {
            method: 'GET',
            headers: {
                'Authorization': 'Basic '+btoa(this.state.login + ':' + this.state.password)
            }
        })
            .then((data) => {
                if (data.status === 401) {
                    console.log("czterysta jeden");
                    return [];
                } else {
                    return data.json();
                }
            })
            .then((data) => {
                console.log(data);
                that.setState({lists: data});
                console.log(data);
            });
    }
    submit() {
        var splink = document.getElementById("sp_link").firstChild.value;
        var spid = document.getElementById("sp_id").firstChild.value;
        var form = new FormData();
        form.append("sp_link", splink);
        form.append("sp_id", spid);
        var that = this;
        fetch('/api/managesplists', {
            method: 'POST',
            headers: {
                'Authorization': 'Basic '+btoa(this.state.login + ':' + this.state.password)
            },
            body: form})
            .then((data) => data.json())
            .then((data) => {
                if(data.correct_entries === 0) {
                    that.setState({message: "Nie znaleziono żadnych wyników na Twojej liście. Czy jest ona pusta?"});
                    that.updateSPLists()
                }
                if(data.correct_entries === -1) {
                    that.setState({message: "Wystąpił błąd przy dodawaniu nowej listy."});
                }
                if(data.correct_entries > 0) {
                    that.setState({message: "Dodano listę."});
                    that.updateSPLists()
                }
            });
    }
    render() {
        var that = this;
        var key = 0;
        return <div>
            <Form.Input label="login" type="text" id="login" />
            <Form.Input label="hasło" type="password" id="password" />
            <Form.Button onClick={this.login}>Submit</Form.Button>
            <hr />
            <Form.Input label="SP link" type="text" id="sp_link" defaultValue="http://dokumenty.uw.edu.pl/dziennik/DWChem"/>
            <Form.Input label="SP ID" type="text" id="sp_id" defaultValue="76df18fd-612f-4ff9-b35a-da9b8688b25f" />
            <Form.Button onClick={this.submit}>Submit</Form.Button>

            {this.state.message}
            <Table>
                <Table.Body>
                    <Table.Row>
                        <Table.HeaderCell>Link</Table.HeaderCell>
                        <Table.HeaderCell>SP Id</Table.HeaderCell>
                        <Table.HeaderCell>Ostatnia migracja</Table.HeaderCell>
                    </Table.Row>
                    {that.state.lists.map((splist) => <Table.Row key={key++}>
                        <Table.Cell>{splist.Link}</Table.Cell>
                        <Table.Cell>{splist.SpID}</Table.Cell>
                        <Table.Cell>{splist.LastMigration}</Table.Cell>
                    </Table.Row>
                    )}
                </Table.Body>
            </Table>
        </div>;
    }
}

export default ManageSpLists;
