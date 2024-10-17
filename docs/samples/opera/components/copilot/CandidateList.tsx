'use client';
import { CandidateList, Candidate } from '../../app/types';
import { useGlobalState } from '../GlobalStateProvider';

interface CandidateListViewProps {
    name?: string;
    type: string; // Type of objects to filter
  }

function CandidateListView({ name = "No Name Provided", type }: CandidateListViewProps ) {
    const { objects } = useGlobalState(); // Access objects from global state

    // Filter objects based on the given type

    const candidates = objects.filter((obj) => obj.type === type).map((obj) => ({
        name: obj.name,
        properties: obj.properties,
    }));

    console.log(objects);
    console.log(type);
    console.log(candidates);

    if (candidates.length === 0) {
        return null; // Return null if no candidates are found
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