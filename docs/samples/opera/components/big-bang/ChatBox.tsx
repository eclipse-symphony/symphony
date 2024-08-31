'use client';
import { CoreMessage } from 'ai';
import { useState } from 'react';
import { SymphonyObject } from '../../app/types';
import { useGlobalState} from '../GlobalStateProvider';

function ChatBox() {
    const [input, setInput] = useState('');
    const [messages, setMessages] = useState<CoreMessage[]>([]);
    const { objects, setObjects} = useGlobalState();

    return (
        <div className="bb_viewport flex flex-col h-full">
            <div className="chat_list flex-grow overflow-auto p-2">
            {messages.map((message, index) => (
                <div key={`${message.role}-${index}`}>
                    {typeof message.content === 'string'
                        ? message.content
                        : message.content
                        .filter(part => part.type === 'text')
                        .map((part, partIndex) => (
                            <div key={partIndex}>{part.text}</div>
                        ))}
                </div>
            ))}
          </div>
          <input
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