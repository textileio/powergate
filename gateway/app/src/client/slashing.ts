import {grpc} from '@improbable-eng/grpc-web'
import {APIClient} from '../_proto/slashing_pb_service'
import {GetRequest, Index} from '../_proto/slashing_pb'

export default class Slashing {

  private client: APIClient

  constructor(serviceHost: string, options?: grpc.RpcOptions) {
    this.client = new APIClient(serviceHost, options)
  }

  get() {
    return new Promise<Index.AsObject>((resolve, reject) => {
      this.client.get(new GetRequest(), (error, resp) => {
        if (error) {
          reject(error)
        } else {
          const val = resp?.getIndex()?.toObject()
          if (!val) {
            reject(new Error('no response object'))
          } else {
            resolve(val)
          }
        }
      })
    })
  }
}