import React from 'react'
import logo from './logo.svg'
import './App.css'

import {grpc} from '@improbable-eng/grpc-web'
import {API} from './_proto/ask_pb_service'
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
  // const c = new APIClient("40.117.82.59:6002")
  const r = new GetRequest()

  grpc.unary(API.Get, {
    request: r,
    host: "40.117.82.59:6002",
    onEnd: res => {
      const { status, statusMessage, headers, message, trailers } = res;
      console.log("getBook.onEnd.status", status, statusMessage);
      console.log("getBook.onEnd.headers", headers);
      if (status === grpc.Code.OK && message) {
        console.log("getBook.onEnd.message", message.toObject());
      }
      console.log("getBook.onEnd.trailers", trailers);
    }
  })

  // const resp = c.get(r, (err, resp) => {
  //   if (err) {
  //     console.log("got error: ", err)
  //   }
  //   if (resp) {
  //     console.log("got response: ", resp)
  //   }
  // })
}

getAsks()

export default App
