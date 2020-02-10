import {grpc} from '@improbable-eng/grpc-web'

// export function unary<
//   TRequest extends grpc.ProtobufMessage, 
//   TResponse extends grpc.ProtobufMessage, 
//   M extends grpc.UnaryMethodDefinition<TRequest, TResponse>>
// (methodDescriptor: M, props: Omit<grpc.UnaryRpcOptions<TRequest, TResponse>, 'onEnd'>) {
//   return new Promise<typeof methodDescriptor.responseType>((resolve, reject) => {
//     const finalProps = {
//       ...props,
//       onEnd: (res: grpc.UnaryOutput<TResponse>) => {
//         const { status, statusMessage, message } = res
//         if (status === grpc.Code.OK) {
//           if (message) {
//             const foo = message.toObject()
//             resolve(message)
//           } else {
//             resolve()
//           }
//         } else {
//           reject(new Error(statusMessage))
//         }
//       }
//     }
//     grpc.unary(methodDescriptor, finalProps)
//   })
// }
