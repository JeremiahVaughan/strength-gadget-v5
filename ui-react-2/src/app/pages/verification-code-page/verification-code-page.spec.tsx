import { render } from '@testing-library/react';

import VerificationCodePage from './verification-code-page';

describe('VerificationCodePage', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<VerificationCodePage />);
    expect(baseElement).toBeTruthy();
  });
});
