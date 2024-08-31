'use client';
import { CoreMessage } from 'ai';
import { useState } from 'react';


function ChatBox() {
    const [input, setInput] = useState('');
    const [messages, setMessages] = useState<CoreMessage[]>([]);

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

                    setMessages(currentMessages => [
                        ...currentMessages,
                        ...newMessages,
                    ]);
                }
            }} />
        </div>
    );
}

export default ChatBox;