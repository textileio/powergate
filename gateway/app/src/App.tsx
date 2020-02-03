import React from 'react'
import logo from './logo.svg'
import './App.css'

import {grpc} from '@improbable-eng/grpc-web'
import {APIClient} from './_proto/ask_pb_service'
import {GetRequest} from './_proto/ask_pb'

const App = () => {
  return (
    <div className="App">
      <header className="App-header">
        <img src={logo} className="App-logo" alt="logo" />
        <p>
          Edit <code>src/App.tsx</code> and save to reload.
        </p>
        <a
          className="App-link"
          href="https://reactjs.org"
          target="_blank"
          rel="noopener noreferrer"
        >
          Learn React
        </a>
      </header>
    </div>
  )
}

function getAsks() {
  const c = new APIClient("40.117.82.59:5002")
  const r = new GetRequest()
  const resp = c.get(r, (err, resp) => {
    if (err) {
      console.log("got error: ", err)
    }
    if (resp) {
      console.log("got response: ", resp)
    }
  })
}

getAsks()

export default App
