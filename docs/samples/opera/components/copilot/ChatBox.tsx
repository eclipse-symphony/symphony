'use client';
import { CoreMessage } from 'ai';
import { useState } from 'react';
import { SymphonyObject } from '../../app/types';
import { useGlobalState} from '../GlobalStateProvider';

function ChatBox() {
    const [input, setInput] = useState('');
    const [messages, setMessages] = useState<CoreMessage[]>([
        { role: 'system', content: 'Symphony defines these object type: a solution that represents a payload, such as an application; a target that represents an endpoint to where a payload can be deployed, such as a server, a Kubernetes cluster, a tiny edge device, or anything that implements Symphony\'s target provider interface; an instance that maps a solution to one or multiple targets; a campaign that defines a workflow, such as canary deployment of an instance; a catalog object that can be used to capture an arbitrary piece of information, such as templates and configurations; a site that is represented by a Symphony control plane, usually refer to a physical location such as an office or a factory floor. Multiple Symphony sites can be linked together to form a tree of control planes. Return your response as an JSON object with a text property that contains a brief explanation of what you are returning. In addition, you can return a list of objects that contains the Symphony objects required to fulfill user\'s requirement. Each object contains a "type", which can be "solution", "target" etc., a "name", and a properties key-value collection. Now my requirement is: ' },
    ]);
    const { objects, setObjects} = useGlobalState();

    return (
        <div className="bb_viewport flex flex-col h-full bg-black p-4">
            <div className="chat_list flex flex-col overflow-auto p-2 h-full">
            {messages.map((message, index) => (
                message.role === 'user' || message.role === 'assistant' ? (
                    <div key={`${message.role}-${index}`} className="text-white">
                        {message.role === 'user' ? (<span className="text-blue-300 font-bold">You: </span>) : (<span className="text-green-300 font-bold">AI: </span>)}
                        {typeof message.content === 'string'
                            ? message.content
                            : message.content
                            .filter(part => part.type === 'text')
                            .map((part, partIndex) => (
                                <div key={partIndex} className="text-white">{part.text}</div>
                            ))}
                    </div>
                ): null
            ))}
          </div>
          <input
            className="border-2 border-gray-300 p-2"
            value={input}
            onChange={event => {
                setInput(event.target.value);
            }}
            onKeyDown={async event => {
                if (event.key === 'Enter') {
                    setMessages(currentMessages => [
                        ...currentMessages,
                        { role: 'user', content: input },
                    ]);

                    const response = await fetch('/api/openai', {
                        method: 'POST',
                        body: JSON.stringify({
                            messages: [...messages, { role: 'user', content: input }],
                        }),
                    });

                    const { messages: newMessages } = await response.json();

                   // Process each message in the newMessages array
                    newMessages.forEach((msg: CoreMessage) => {
                        if (msg.role === 'assistant' && Array.isArray(msg.content)) {
                        // Find the part of the content that is text and contains the JSON string
                        const textPart = msg.content.find(
                            (part: any) => part.type === 'text' && part.text
                        );

                        if (textPart) {
                            try {
                                // Parse the JSON string contained in the text part
                                const parsedContent = JSON.parse(textPart.text);

                                // Update the messages with the extracted text
                                if (parsedContent.text) {
                                    setMessages((currentMessages) => [
                                    ...currentMessages,
                                    { role: 'assistant', content: parsedContent.text },
                                ]);
                                if (parsedContent.objects) {
                                    setObjects(parsedContent.objects);
                                }
                            }                     
                        } catch (error) {
                            console.error('Failed to parse AI message content as JSON:', error);
                            setMessages((currentMessages) => [
                                ...currentMessages,
                                { role: 'assistant', content: textPart.text },
                            ]);
                        }
                    }
                    setInput(''); // Clear the input after sending
                } else {
                  // If it's not an AI message or not in the expected format, just add it to the list
                  setMessages((currentMessages) => [...currentMessages, msg]);
                }
              });

              setInput(''); // Clear the input after sending
                }
            }} />
        </div>
    );
}

export default ChatBox;