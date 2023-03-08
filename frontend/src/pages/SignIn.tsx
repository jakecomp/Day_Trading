/* eslint-disable prettier/prettier */
import { BigBlackButton } from '../components/atoms/button'
import { SignBackground } from '../components/sign_in/background'
import { Card } from '../components/sign_in/card'
import { SignInHeader, SignInLink, SignInText } from '../components/sign_in/text'
import { InputContainer, LinkContainer } from '../components/sign_in/containers'
import { FieldForm, InputLabel, SignField } from '../components/sign_in/field'
import { useForm } from 'react-hook-form'
import { SimpleLink } from '../components/atoms/links'
import { Link } from 'react-router-dom'

export const SignIn = () => {
    interface signForm {
        username: string
        password: string
    }

    let socket: WebSocket

    const { register, handleSubmit } = useForm<signForm>({ mode: 'onSubmit' })
    const RetrieveData = (data: signForm) => {
        const report = {
            username: data.username,
            password: data.password,
        }

        try {
            fetch('http://10.9.0.4:8000/signin', {
                method: 'POST',
                headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
                body: JSON.stringify(report),
            })
                .then((response) => response.text())
                .then((response) => {
                    socket = new WebSocket('ws://10.9.0.4:8000/ws?token=' + response)
                    socket.onopen = function () {
                        socket.send('Hi Hi Server')
                        socket.onmessage = (msg: any) => {
                            console.log('Server Message: ' + msg.data)
                        }
                    }
                })
        } catch (error) {
            console.error(error)
        }
    }

    return (
        <div>
            <SignBackground>
                <Card>
                    <SignInHeader>Sign In to Day-trades</SignInHeader>
                    <FieldForm onSubmit={handleSubmit(RetrieveData)}>
                        <InputContainer>
                            <InputLabel>Username</InputLabel>
                            <SignField {...register('username')}></SignField>
                            <InputLabel>Password</InputLabel>
                            <SignField type='password' {...register('password')}></SignField>{' '}
                        </InputContainer>
                    <SimpleLink to='/home'><BigBlackButton>Sign in</BigBlackButton> </SimpleLink>
                    </FieldForm>
                    <LinkContainer>
                        <SignInText>New to Day-trades?</SignInText>
                        <SignInLink to='/'>Create an account</SignInLink>
                    </LinkContainer>
                </Card>
            </SignBackground>
        </div>
    )
}
