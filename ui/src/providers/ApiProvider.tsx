import type { ReactNode } from 'react';

interface ApiProviderProps {
    children: ReactNode;
}

export function ApiProvider({ children }: ApiProviderProps) {
    // OSS version doesn't need authentication setup
    return <>{children}</>;
}