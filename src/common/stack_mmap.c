/**
 * (C) Copyright 2016-2021 Intel Corporation.
 *
 * SPDX-License-Identifier: BSD-2-Clause-Patent
 */
/**
 * \file
 *
 * This file is part of the DAOS server. It implements an alternate way
 * for ULTs stacks allocation, based on mmap() of MAP_STACK|MAP_GROWSDOWN
 * regions, in order to allow overrun detection along with automatic growth
 * possibility.
 */

#ifdef ULT_MMAP_STACK
#include <daos/common.h>
#include <daos/stack_mmap.h>
#include <errno.h>
#include <string.h>

/* ABT_key for mmap()'ed ULT stacks */
ABT_key stack_key;

/* pool of free stacks */
pthread_mutex_t stack_free_list_lock;
d_list_t stack_free_list;

/* callback to free stack upon ULT exit and stack_key deregister */
void free_stack(void *arg)
{
	mmap_stack_desc_t *desc = (mmap_stack_desc_t *)arg;

	D_MUTEX_LOCK(&stack_free_list_lock);
	d_list_add_tail(&stack_free_list, &desc->stack_list);
	D_MUTEX_UNLOCK(&stack_free_list_lock);
	D_DEBUG(DB_MEM, "%p mmap()'ed stack has been put on stack free list.\n",
		desc->stack);
}

/* wrapper for ULT main function, mainly to register mmap()'ed stack
 * descriptor as ABT_key to ensure stack munmap() upon ULT exit
 */
static void mmap_stack_wrapper(void *arg)
{
	mmap_stack_desc_t *desc = (mmap_stack_desc_t *)arg;

	ABT_key_set(stack_key, desc);
	desc->thread_func(desc->thread_arg);
}

/* XXX
 * presently ABT_thread_create_many() is not used in DAOS code, but if it
 * becomes we will also have to introduce a corresponding wrapper
 */

int mmap_stack_thread_create(ABT_pool pool, void (*thread_func)(void *),
			     void *thread_arg, ABT_thread_attr attr,
			     ABT_thread *newthread)
{
	ABT_thread_attr new_attr = ABT_THREAD_ATTR_NULL;
	int rc;
	void *stack;
	mmap_stack_desc_t *mmap_stack_desc = NULL;
	size_t stack_size = MMAPED_ULT_STACK_SIZE, new_stack_size;

	if (attr != ABT_THREAD_ATTR_NULL) {
		ABT_thread_attr_get_stack(attr, &stack, &stack_size);
		if (stack != NULL) {
			/* an other external stack allocation method is being
			 *  used, nothing to do
			 */
			rc = ABT_thread_create(pool, thread_func, thread_arg,
					       attr, newthread);
			return rc;
		}
		if (stack_size < MMAPED_ULT_STACK_SIZE)
			stack_size = MMAPED_ULT_STACK_SIZE;
	} else {
		rc = ABT_thread_attr_create(&new_attr);
		if (rc != ABT_SUCCESS) {
			D_ERROR("Create ABT thread attr failed: %d\n", rc);
			return rc;
		}
		attr = new_attr;
	}

	D_MUTEX_LOCK(&stack_free_list_lock);
	if (!d_list_empty(&stack_free_list)) {
		mmap_stack_desc = container_of(stack_free_list.next,
					       mmap_stack_desc_t, stack_list);
		d_list_del_init(stack_free_list.next);
		D_MUTEX_UNLOCK(&stack_free_list_lock);
		stack = mmap_stack_desc->stack;
		/* XXX may need to reevaluate stack size since a growth may
		 * have occured during previous context life time ?
		 */
		stack_size = mmap_stack_desc->stack_size;
		D_DEBUG(DB_MEM, "%p mmap()'ed stack has been taken from stack free list.\n",
			mmap_stack_desc->stack);
	} else {
		D_MUTEX_UNLOCK(&stack_free_list_lock);
		stack = mmap(NULL, stack_size, PROT_READ | PROT_WRITE,
			     MAP_PRIVATE | MAP_ANONYMOUS | MAP_STACK | MAP_GROWSDOWN,
			     -1, 0);
		if (stack == MAP_FAILED) {
			D_ERROR("Failed to mmap() ULT stack : %s\n", strerror(errno));
			/* return an ABT error */
			D_GOTO(out_err, rc = ABT_ERR_MEM);
		}

		/* put descriptor at bottom of mmap()'ed stack */
		mmap_stack_desc = (mmap_stack_desc_t *)(stack + stack_size -
				  sizeof(mmap_stack_desc_t));

		/* start to fill descriptor */
		mmap_stack_desc->stack = stack;
		mmap_stack_desc->stack_size = stack_size;
		D_INIT_LIST_HEAD(&mmap_stack_desc->stack_list);
		D_DEBUG(DB_MEM, "%p mmap()'ed stack has been taken allocated.\n",
			mmap_stack_desc->stack);
	}

	/* continue to fill/update descriptor */
	mmap_stack_desc->thread_func = thread_func;
	mmap_stack_desc->thread_arg = thread_arg;

	/* usable stack size */
	new_stack_size = stack_size - sizeof(mmap_stack_desc_t);

	rc = ABT_thread_attr_set_stack(attr, stack, new_stack_size);
	if (rc != ABT_SUCCESS) {
		D_ERROR("Failed to set stack attrs : %d\n", rc);
		D_GOTO(out_err, rc);
	}

	/* XXX if newthread is set, we may need to use
	 * ABT_thread_set_specific() ??
	 */
	rc = ABT_thread_create(pool, mmap_stack_wrapper, mmap_stack_desc, attr,
			       newthread);
	if (rc != ABT_SUCCESS) {
		D_ERROR("Failed to create ULT : %d\n", rc);
		D_GOTO(out_err, rc);
	}
out_err:
	if (rc && mmap_stack_desc != NULL)
		free_stack(mmap_stack_desc);
	/* free local attr if used */
	if (new_attr != ABT_THREAD_ATTR_NULL)
		ABT_thread_attr_free(&new_attr);
	return rc;
}

int mmap_stack_thread_create_on_xstream(ABT_xstream xstream,
					void (*thread_func)(void *),
					void *thread_arg, ABT_thread_attr attr,
					ABT_thread *newthread)
{
	ABT_thread_attr new_attr = ABT_THREAD_ATTR_NULL;
	int rc;
	void *stack;
	mmap_stack_desc_t *mmap_stack_desc = NULL;
	size_t stack_size = MMAPED_ULT_STACK_SIZE, new_stack_size;

	if (attr != ABT_THREAD_ATTR_NULL) {
		ABT_thread_attr_get_stack(attr, &stack, &stack_size);
		if (stack != NULL) {
			/* an other external stack allocation method is being
			 *  used, nothing to do
			 */
			rc = ABT_thread_create_on_xstream(xstream, thread_func,
							  thread_arg, attr,
							  newthread);
			return rc;
		}
		if (stack_size < MMAPED_ULT_STACK_SIZE)
			stack_size = MMAPED_ULT_STACK_SIZE;
	} else {
		rc = ABT_thread_attr_create(&new_attr);
		if (rc != ABT_SUCCESS) {
			D_ERROR("Create ABT thread attr failed: %d\n", rc);
			return rc;
		}
	}

	D_MUTEX_LOCK(&stack_free_list_lock);
	if (!d_list_empty(&stack_free_list)) {
		mmap_stack_desc = container_of(stack_free_list.next,
					       mmap_stack_desc_t, stack_list);
		d_list_del_init(stack_free_list.next);
		D_MUTEX_UNLOCK(&stack_free_list_lock);
		stack = mmap_stack_desc->stack;
		/* XXX may need to reevaluate stack size since a growth may
		 * have occured during previous context life time ?
		 */
		stack_size = mmap_stack_desc->stack_size;
		D_DEBUG(DB_MEM, "%p mmap()'ed stack has been taken from stack free list.\n",
			mmap_stack_desc->stack);
	} else {
		D_MUTEX_UNLOCK(&stack_free_list_lock);
		stack = mmap(NULL, stack_size, PROT_READ | PROT_WRITE,
			     MAP_PRIVATE | MAP_ANONYMOUS | MAP_STACK | MAP_GROWSDOWN,
			     -1, 0);
		if (stack == MAP_FAILED) {
			D_ERROR("Failed to mmap() ULT stack : %s\n", strerror(errno));
			/* return an ABT error */
			D_GOTO(out_err, rc = ABT_ERR_MEM);
		}

		/* put descriptor at bottom of mmap()'ed stack */
		mmap_stack_desc = (mmap_stack_desc_t *)(stack + stack_size -
				  sizeof(mmap_stack_desc_t));

		/* start to fill descriptor */
		mmap_stack_desc->stack = stack;
		mmap_stack_desc->stack_size = stack_size;
		D_INIT_LIST_HEAD(&mmap_stack_desc->stack_list);
		D_DEBUG(DB_MEM, "%p mmap()'ed stack has been taken allocated.\n",
			mmap_stack_desc->stack);
	}

	/* continue to fill/update descriptor */
	mmap_stack_desc->thread_func = thread_func;
	mmap_stack_desc->thread_arg = thread_arg;

	/* usable stack size */
	new_stack_size = stack_size - sizeof(mmap_stack_desc_t);

	rc = ABT_thread_attr_set_stack(attr, stack, new_stack_size);
	if (rc != ABT_SUCCESS) {
		D_ERROR("Failed to set stack attrs : %d\n", rc);
		D_GOTO(out_err, rc);
	}

	/* XXX if newthread is set, we may need to use
	 * ABT_thread_set_specific() ??
	 */
	rc = ABT_thread_create_on_xstream(xstream, mmap_stack_wrapper,
					  mmap_stack_desc, attr, newthread);
	if (rc != ABT_SUCCESS) {
		D_ERROR("Failed to create ULT : %d\n", rc);
		D_GOTO(out_err, rc);
	}
out_err:
	if (rc && mmap_stack_desc != NULL)
		free_stack(mmap_stack_desc);
	if (new_attr != ABT_THREAD_ATTR_NULL)
		ABT_thread_attr_free(&attr);
	return rc;
}
#endif
