import React from 'react'
import CopilotCanvas from '@/components/copilot/CopilotCanvas';

async function CopilotPage() {
    return (
        <div className="flex flex-col border-gray-200 bg-gray-100 m-5 p-5 h-[83%]">
            <CopilotCanvas />
        </div>
    );
}

export default CopilotPage;