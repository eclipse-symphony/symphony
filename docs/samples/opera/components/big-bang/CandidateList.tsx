'use client';
import { CandidateList, Candidate } from '../../app/types';

function CandidateListView({ name = "No Name Provided", candidates = [] }: Partial<CandidateList>) {
    if (!candidates || candidates.length === 0) {
        return null;
    }
    
    return (
        <div className="candidate_list">
            <h1 className="candidate_title">{name}</h1>
            {candidates.length > 0 ? (
                <ul>
                    {candidates.map((candidate, index) => (
                        <li key={index}>{candidate.name}</li>
                    ))}
                </ul>
            ) : (
                <p className="text-gray-500">No candidates available.</p> 
            )}
        </div>
    );
}


export default CandidateListView;